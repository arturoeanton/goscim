package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/arturoeanton/goscim/scim"
	"github.com/gin-gonic/gin"
)

// Version is the release this binary was built from.
const Version = "1.0.0"

// Server timeouts. net/http applies none by default, which leaves a connection
// that dribbles a request out one byte at a time holding a goroutine
// indefinitely.
const (
	readHeaderTimeout = 10 * time.Second
	readTimeout       = 30 * time.Second
	writeTimeout      = 30 * time.Second
	idleTimeout       = 120 * time.Second
	shutdownTimeout   = 20 * time.Second
	maxHeaderBytes    = 1 << 20 // 1 MiB
)

func main() {
	log.SetFlags(log.LstdFlags | log.LUTC)
	log.Printf("GoSCIM %s starting", Version)

	authenticator, err := scim.NewAuthenticatorFromEnv()
	if err != nil {
		log.Fatalln(">>>", err.Error())
	}
	if _, anonymous := authenticator.(*scim.AnonymousAuthenticator); anonymous {
		log.Println("WARNING: SCIM_AUTH=none - every request is served unauthenticated")
	}

	folderConfig := envOr("SCIM_CONFIG_DIR", "config")
	// Bucket settings are read relative to the same directory, so the process
	// no longer has to be started from the repository root.
	scim.FolderBucketSetting = strings.TrimSuffix(folderConfig, "/") + "/bucketSettings/"

	if err := scim.InitDB(); err != nil {
		log.Fatalln(">>>", err.Error())
	}
	defer scim.DB.Close()

	r := gin.Default()
	if err := r.SetTrustedProxies(trustedProxies()); err != nil {
		log.Fatalln(">>>", err.Error())
	}
	if _, err := scim.NewRouter(folderConfig, r, authenticator); err != nil {
		log.Fatalln(">>>", err.Error())
	}

	server := &http.Server{
		Addr:              envOr("SCIM_PORT", ":8080"),
		Handler:           r,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		MaxHeaderBytes:    maxHeaderBytes,
	}

	if err := run(server); err != nil {
		log.Fatalln(">>>", err.Error())
	}
	log.Println("stopped")
}

// run serves until the process is asked to stop, then drains in-flight
// requests before returning. Without this a deployment rolling the process
// cuts every request that happens to be in progress.
func run(server *http.Server) error {
	stopping := make(chan os.Signal, 1)
	signal.Notify(stopping, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(stopping)

	failed := make(chan error, 1)
	go func() {
		log.Printf("listening on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			failed <- err
			return
		}
		failed <- nil
	}()

	select {
	case err := <-failed:
		return err
	case signal := <-stopping:
		log.Printf("%s received, draining for up to %s", signal, shutdownTimeout)
	}

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		return err
	}
	return <-failed
}

func envOr(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}

// trustedProxies decides whose X-Forwarded-For gin will believe. The default
// stays loopback only; a deployment behind a load balancer has to name it.
func trustedProxies() []string {
	raw := strings.TrimSpace(os.Getenv("SCIM_TRUSTED_PROXIES"))
	if raw == "" {
		return []string{"127.0.0.1"}
	}
	if strings.EqualFold(raw, "none") {
		return nil
	}
	proxies := make([]string, 0)
	for _, proxy := range strings.Split(raw, ",") {
		if trimmed := strings.TrimSpace(proxy); trimmed != "" {
			proxies = append(proxies, trimmed)
		}
	}
	return proxies
}

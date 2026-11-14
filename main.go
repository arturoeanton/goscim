package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	configembed "github.com/arturoeanton/goscim/config"
	"github.com/arturoeanton/goscim/scim"
	"github.com/gin-gonic/gin"
)

// Version is the release this binary was built from. Can be overridden at
// link time with -ldflags="-X main.Version=...".
var Version = "1.4.0"

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
	demoDefaultToken  = "demo-secret-token"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "conformance":
			os.Exit(scim.RunConformanceCLI(os.Args[2:], os.Stdout, os.Stderr))
		case "version":
			fmt.Printf("GoSCIM %s\n", Version)
			return
		}
	}

	healthcheck := flag.Bool("healthcheck", false, "probe /readyz on the configured SCIM_PORT and exit")
	demoMode := flag.Bool("demo", false, "start instantly with in-memory store, embedded HTMX UI (/ui), sample data, and bearer auth")
	enableUI := flag.Bool("ui", false, "mount the embedded HTMX management and live inspector console at /ui")
	storeFlag := flag.String("store", "", "storage backend override: couchbase, sqlite, or memory")
	flag.Parse()

	if *healthcheck {
		os.Exit(runHealthcheck(envOr("SCIM_PORT", ":8080")))
	}

	if *storeFlag != "" {
		_ = os.Setenv("SCIM_STORE", *storeFlag)
	}
	if *enableUI {
		_ = os.Setenv("SCIM_UI", "true")
	}
	if *demoMode {
		applyDemoDefaults()
	}

	scim.SetFallbackConfigFS(configembed.FS)
	scim.DefaultMetrics.SetBuildVersion(Version)

	log.SetFlags(log.LstdFlags | log.LUTC)
	log.Printf("GoSCIM %s starting", Version)

	authenticator, err := scim.NewAuthenticatorFromEnv()
	if err != nil {
		log.Fatalln(">>>", err.Error())
	}
	if _, anonymous := authenticator.(*scim.AnonymousAuthenticator); anonymous {
		log.Println("WARNING: SCIM_AUTH=none - every request is served unauthenticated")
	}

	passwordMode, err := scim.PasswordModeFromEnv()
	if err != nil {
		log.Fatalln(">>>", err.Error())
	}
	if passwordMode == scim.PasswordModePlain {
		log.Println("WARNING: SCIM_PASSWORD_MODE=plain - passwords are stored unhashed")
	}

	if _, err := scim.ConfigureAuditLoggerFromEnv(nil); err != nil {
		log.Fatalln(">>>", err.Error())
	}

	webhooks, err := scim.ConfigureWebhooksFromEnv()
	if err != nil {
		log.Fatalln(">>>", err.Error())
	}
	if webhooks != nil {
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = webhooks.Close(ctx)
		}()
	}

	folderConfig := envOr("SCIM_CONFIG_DIR", "config")
	// Bucket settings are read relative to the same directory, so the process
	// no longer has to be started from the repository root.
	scim.FolderBucketSetting = strings.TrimSuffix(folderConfig, "/") + "/bucketSettings/"

	backend, err := scim.InitStoreFromEnv()
	if err != nil {
		log.Fatalln(">>>", err.Error())
	}
	defer scim.DB.Close()
	log.Printf("store backend initialised: %s", backend)

	r := gin.Default()
	if err := r.SetTrustedProxies(trustedProxies()); err != nil {
		log.Fatalln(">>>", err.Error())
	}
	if _, err := scim.NewRouter(folderConfig, r, authenticator); err != nil {
		log.Fatalln(">>>", err.Error())
	}

	if scim.UIEnabledFromEnv() {
		r.GET("/", func(c *gin.Context) {
			c.Redirect(http.StatusSeeOther, "/ui")
		})
	}

	listenAddr := envOr("SCIM_PORT", ":8080")
	if *demoMode {
		seeded, err := scim.SeedDemoResources("", false)
		if err != nil {
			log.Printf("warning: seeding demo resources: %v", err)
		}
		printDemoBanner(listenAddr, seeded)
	}

	server := &http.Server{
		Addr:              listenAddr,
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

func applyDemoDefaults() {
	setIfEmpty("GIN_MODE", "release")
	setIfEmpty("SCIM_STORE", "memory")
	setIfEmpty("SCIM_UI", "true")
	setIfEmpty("SCIM_AUTH", "bearer")
	if os.Getenv("SCIM_BEARER_TOKEN") == "" && os.Getenv("SCIM_BEARER_TOKENS") == "" {
		_ = os.Setenv("SCIM_BEARER_TOKEN", demoDefaultToken)
		_ = os.Setenv("SCIM_BEARER_SUBJECT", "demo-admin")
		_ = os.Setenv("SCIM_BEARER_ROLES", "admin,superadmin,user,role1,role2")
	}
}

func setIfEmpty(key, value string) {
	if strings.TrimSpace(os.Getenv(key)) == "" {
		_ = os.Setenv(key, value)
	}
}

func printDemoBanner(listenAddr string, seeded int) {
	host, port, err := net.SplitHostPort(listenAddr)
	if err != nil {
		port = strings.TrimPrefix(listenAddr, ":")
	}
	if host == "" || host == "0.0.0.0" || host == "::" {
		host = "127.0.0.1"
	}
	base := "http://" + net.JoinHostPort(host, port)
	token := envOr("SCIM_BEARER_TOKEN", demoDefaultToken)

	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "========================================================================\n")
	fmt.Fprintf(os.Stderr, "  GoSCIM %s (Demo Mode - %d sample resources ready)\n", Version, seeded)
	fmt.Fprintf(os.Stderr, "------------------------------------------------------------------------\n")
	fmt.Fprintf(os.Stderr, "  Web Console (HTMX):  %s/ui\n", base)
	fmt.Fprintf(os.Stderr, "  SCIM 2.0 Base URL:   %s/scim/v2\n", base)
	fmt.Fprintf(os.Stderr, "  Multi-Tenant URL:    %s/tenants/acme/scim/v2\n", base)
	fmt.Fprintf(os.Stderr, "  Prometheus Metrics:  %s/metrics\n", base)
	fmt.Fprintf(os.Stderr, "  Bearer Token:        %s\n", token)
	fmt.Fprintf(os.Stderr, "------------------------------------------------------------------------\n")
	fmt.Fprintf(os.Stderr, "  Try it:\n")
	fmt.Fprintf(os.Stderr, "    curl -s -H \"Authorization: Bearer %s\" %s/scim/v2/Users | jq .\n", token, base)
	fmt.Fprintf(os.Stderr, "    goscim conformance -url %s/scim/v2 -token %s\n", base, token)
	fmt.Fprintf(os.Stderr, "========================================================================\n\n")
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

// runHealthcheck probes the local /readyz endpoint so distroless containers
// need neither a shell nor curl for Docker/Compose healthchecks.
func runHealthcheck(listenAddr string) int {
	host, port, err := net.SplitHostPort(listenAddr)
	if err != nil {
		port = strings.TrimPrefix(listenAddr, ":")
	}
	if host == "" || host == "0.0.0.0" || host == "::" {
		host = "127.0.0.1"
	}
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://" + net.JoinHostPort(host, port) + "/readyz")
	if err != nil {
		return 1
	}
	_ = resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return 0
	}
	return 1
}

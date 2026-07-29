package main

import (
	"log"
	"os"

	"github.com/arturoeanton/goscim/scim"
	"github.com/gin-gonic/gin"
)

func main() {

	authenticator, err := scim.NewAuthenticatorFromEnv()
	if err != nil {
		log.Fatalln(">>>", err.Error())
	}
	if _, anonymous := authenticator.(*scim.AnonymousAuthenticator); anonymous {
		log.Println("WARNING: SCIM_AUTH=none - every request is served unauthenticated")
	}

	if err := scim.InitDB(); err != nil {
		log.Fatalln(">>>", err.Error())
	}
	defer scim.DB.Close()

	log.Println("GoScim v0.1")
	folderConfig := "config"

	port := os.Getenv("SCIM_PORT")
	if port == "" {
		port = ":8080"
	}

	r := gin.Default()
	r.SetTrustedProxies([]string{"127.0.0.1"})
	if _, err := scim.NewRouter(folderConfig, r, authenticator); err != nil {
		log.Fatalln(">>>", err.Error())
	}
	r.Run(port)
}

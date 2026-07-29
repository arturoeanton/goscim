package main

import (
	"log"
	"os"

	"github.com/arturoeanton/goscim/scim"
	"github.com/gin-gonic/gin"
)

func main() {

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
	if _, err := scim.NewRouter(folderConfig, r); err != nil {
		log.Fatalln(">>>", err.Error())
	}
	r.Run(port)
}

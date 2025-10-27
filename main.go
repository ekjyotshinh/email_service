package main

import (
	"log"
	"os"

	"github.com/ekjyotshinh/email-service/controller"
	"github.com/ekjyotshinh/email-service/db"
	"github.com/ekjyotshinh/email-service/worker"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	checkEnvVars()

	db.Connect()
	worker.Start()

	r := gin.Default()
	r.POST("/email", controller.SendEmail)
	r.GET("/emails", controller.GetEmails)

	r.Run(":8080")
}

func checkEnvVars() {
	envars := []string{
		"DB_HOST",
		"DB_USER",
		"DB_PASS",
		"DB_NAME",
		"SMTP_HOST",
		"SMTP_PORT",
		"SMTP_USERNAME",
		"SMTP_PASSWORD",
	}

	for _, envvar := range envars {
		if os.Getenv(envvar) == "" {
			log.Fatalf("Error: %s environment variable not set", envvar)
		}
	}
}
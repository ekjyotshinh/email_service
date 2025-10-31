package main

import (
	"github.com/ekjyotshinh/email-service/config"
	"github.com/ekjyotshinh/email-service/controller"
	"github.com/ekjyotshinh/email-service/db"
	"github.com/ekjyotshinh/email-service/worker"
	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadConfig()

	db.Connect()
	worker.Start()

	r := gin.Default()
	r.POST("/email", controller.SendEmail)
	r.GET("/emails", controller.GetEmails)
	r.GET("/health", controller.HealthCheck)

	r.Run(":8080")
}
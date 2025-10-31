package controller

import (
	"net/http"
	"time"

	"github.com/ekjyotshinh/email-service/db"
	"github.com/ekjyotshinh/email-service/model"
	"github.com/gin-gonic/gin"
)

func SendEmail(c *gin.Context) {
	var email model.Email
	if err := c.BindJSON(&email); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if email.SendTime.IsZero() || email.SendTime.Before(time.Now()) {
		email.SendTime = time.Now()
	}

	email.Status = db.StatusPending
	db.InsertEmail(&email)
	c.JSON(http.StatusOK, gin.H{"message": "Email queued"})
}

func GetEmails(c *gin.Context) {
	emails, err := db.GetEmails()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, emails)
}

func HealthCheck(c *gin.Context) {
	err := db.DB.Ping()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "unhealthy"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}
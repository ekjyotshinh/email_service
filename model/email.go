package model
import (
	"time"
)

type Email struct {
	ID       int       `json:"id"`
	To       string    `json:"to"`
	Subject  string    `json:"subject"`
	Body     string    `json:"body"`
	Status   string    `json:"status"`
	RetryCount int       `json:"retry_count"`
	SendTime time.Time `json:"send_time"`
}

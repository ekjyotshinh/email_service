package worker

import (
	"log"
	"math"
	"time"

	"github.com/ekjyotshinh/email-service/db"
	"github.com/ekjyotshinh/email-service/service"
)

const (
	poolInterval    = 5 * time.Second
	maxWorkers      = 5
	maxRetries      = 3
	baseBackoffTime = 1 * time.Minute
)

func Start() {
	for i := 0; i < maxWorkers; i++ {
		go worker()
	}
}

func worker() {
	for {
		emails, err := db.GetPendingEmailsAndMarkAsProcessing(maxWorkers)
		if err != nil {
			log.Printf("Error getting pending emails: %v", err)
			time.Sleep(poolInterval)
			continue
		}

		for _, email := range emails {
			err := service.SendEmail(email)
			if err != nil {
				log.Printf("ERROR: failed to send email. ID: %d, To: %s, Error: %v", email.ID, email.To, err)
				if email.RetryCount < maxRetries {
					newRetryCount := email.RetryCount + 1
					backoffDuration := time.Duration(math.Pow(2, float64(newRetryCount))) * baseBackoffTime
					nextSendTime := time.Now().Add(backoffDuration)
					err := db.IncrementRetryCount(email.ID, newRetryCount, nextSendTime)
					if err != nil {
						log.Printf("ERROR: failed to increment retry count for email ID: %d, Error: %v", email.ID, err)
					}
				} else {
					err := db.MarkAsFailed(email.ID)
					if err != nil {
						log.Printf("ERROR: failed to mark email as failed. ID: %d, Error: %v", email.ID, err)
					}
				}
			} else {
				log.Printf("INFO: email sent successfully. ID: %d, To: %s", email.ID, email.To)
				db.MarkAsSent(email.ID)
			}
		}

		time.Sleep(poolInterval)
	}
}
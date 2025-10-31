package worker

import (
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/ekjyotshinh/email-service/config"
	"github.com/ekjyotshinh/email-service/db"
	"github.com/ekjyotshinh/email-service/model"
	"github.com/ekjyotshinh/email-service/service"
)

var emailJobChan chan model.Email

func Start() {
	emailJobChan = make(chan model.Email, config.AppConfig.MaxWorkers)
	go dispatcher()
	go recoverStuckProcessingEmails()

	for i := 0; i < config.AppConfig.MaxWorkers; i++ {
		go worker(i, emailJobChan)
	}
}

func dispatcher() {
	for {
		emails, err := db.GetPendingEmailsAndMarkAsProcessing(config.AppConfig.MaxWorkers)
		if err != nil {
			log.Printf("Error getting pending emails: %v", err)
			time.Sleep(config.AppConfig.PoolInterval * time.Second)
			continue
		}

		for _, email := range emails {
			emailJobChan <- email
		}

		if len(emails) == 0 {
			time.Sleep(config.AppConfig.PoolInterval * time.Second)
		}
	}
}

func worker(id int, jobs <-chan model.Email) {
	for email := range jobs {
		log.Printf("Worker %d started processing email ID: %d", id, email.ID)
		err := service.SendEmail(email)
		if err != nil {
			log.Printf("ERROR: worker %d failed to send email. ID: %d, To: %s, Error: %v", id, email.ID, email.To, err)
			if email.RetryCount < config.AppConfig.MaxRetries {
				newRetryCount := email.RetryCount + 1
				backoffDuration := time.Duration(math.Pow(2, float64(newRetryCount)))*config.AppConfig.BaseBackoffTime*time.Second + time.Duration(rand.Intn(1000))*time.Millisecond
				nextSendTime := time.Now().Add(backoffDuration)
				err := db.IncrementRetryCount(email.ID, newRetryCount, nextSendTime)
				if err != nil {
					log.Printf("ERROR: worker %d failed to increment retry count for email ID: %d, Error: %v", id, email.ID, err)
				}
			} else {
				err := db.MarkAsFailed(email.ID)
				if err != nil {
					log.Printf("ERROR: worker %d failed to mark email as failed. ID: %d, Error: %v", id, email.ID, err)
				}
			}
		} else {
			log.Printf("INFO: worker %d email sent successfully. ID: %d, To: %s", id, email.ID, email.To)
			db.MarkAsSent(email.ID)
		}
		log.Printf("Worker %d finished processing email ID: %d", id, email.ID)
	}
}

func recoverStuckProcessingEmails() {
	for {
		time.Sleep(config.AppConfig.StuckProcessingCheck * time.Second)
		log.Println("Running job to recover stuck processing emails...")
		err := db.ResetStuckProcessingEmails(config.AppConfig.ProcessingTimeout * time.Second)
		if err != nil {
			log.Printf("Error recovering stuck processing emails: %v", err)
		}
	}
}
package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/ekjyotshinh/email-service/config"
	"github.com/ekjyotshinh/email-service/model"
	_ "github.com/lib/pq"
)

const (
	StatusPending    = "pending"
	StatusProcessing = "processing"
	StatusSent       = "sent"
	StatusFailed     = "failed"
)

var DB *sql.DB

func Connect() {
	var err error
	DB, err = connectWithRetry()
	if err != nil {
		log.Fatal("DB connection failed after multiple retries: ", err)
	}
	log.Println("Connected to DB successfully")
}

func connectWithRetry() (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		config.AppConfig.DBHost, config.AppConfig.DBUser, config.AppConfig.DBPass, config.AppConfig.DBName)

	var db *sql.DB
	var err error

	for i := 0; i < config.AppConfig.MaxRetries; i++ {
		db, err = sql.Open("postgres", dsn)
		if err != nil {
			log.Printf("DB connection error: %v. Retrying in %v...", err, config.AppConfig.PoolInterval*time.Second)
			time.Sleep(config.AppConfig.PoolInterval * time.Second)
			continue
		}

		err = db.Ping()
		if err != nil {
			log.Printf("DB ping failed: %v. Retrying in %v...", err, config.AppConfig.PoolInterval*time.Second)
			time.Sleep(config.AppConfig.PoolInterval * time.Second)
			continue
		}

		return db, nil
	}

	return nil, fmt.Errorf("failed to connect to the database after %d retries", config.AppConfig.MaxRetries)

}

func InsertEmail(email *model.Email) error {
	err := DB.QueryRow("INSERT INTO emails (to, subject, body, status, send_time, retry_count, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW()) RETURNING id",
		email.To, email.Subject, email.Body, email.Status, email.SendTime, 0).Scan(&email.ID)
	return err
}

func GetPendingEmailsAndMarkAsProcessing(limit int) ([]model.Email, error) {
	tx, err := DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	rows, err := tx.Query(`
		SELECT id, "to", subject, body, retry_count
		FROM emails
		WHERE status = $1 AND send_time <= $2
		ORDER BY send_time
		FOR UPDATE SKIP LOCKED
		LIMIT $3
	`, StatusPending, time.Now(), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var emails []model.Email
	var emailIDs []int
	for rows.Next() {
		var e model.Email
		if err := rows.Scan(&e.ID, &e.To, &e.Subject, &e.Body, &e.RetryCount); err != nil {
			continue
		}
		emails = append(emails, e)
		emailIDs = append(emailIDs, e.ID)
	}

	if len(emailIDs) == 0 {
		return emails, nil
	}

	_, err = tx.Exec(`
		UPDATE emails
		SET status = $1, updated_at = NOW()
		WHERE id = ANY($2)
	`, StatusProcessing, emailIDs)

	if err != nil {
		return nil, err
	}

	return emails, tx.Commit()
}

func IncrementRetryCount(id int, retryCount int, nextSendTime time.Time) error {
	_, err := DB.Exec("UPDATE emails SET retry_count = $1, send_time = $2, status = $3, updated_at = NOW() WHERE id = $4", retryCount, nextSendTime, StatusPending, id)
	return err
}

func MarkAsSent(id int) error {
	_, err := DB.Exec("UPDATE emails SET status = $1, updated_at = NOW() WHERE id = $2", StatusSent, id)
	return err
}

func MarkAsFailed(id int) error {
	_, err := DB.Exec("UPDATE emails SET status = $1, updated_at = NOW() WHERE id = $2", StatusFailed, id)
	return err
}

func GetEmails() ([]model.Email, error) {
	rows, err := DB.Query("SELECT id, \"to\", subject, body, status, send_time, retry_count, created_at, updated_at FROM emails")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var emails []model.Email
	for rows.Next() {
		var e model.Email
		if err := rows.Scan(&e.ID, &e.To, &e.Subject, &e.Body, &e.Status, &e.SendTime, &e.RetryCount, &e.CreatedAt, &e.UpdatedAt); err != nil {
			continue
		}
		emails = append(emails, e)
	}
	return emails, nil
}

func ResetStuckProcessingEmails(timeout time.Duration) error {
	_, err := DB.Exec(`
		UPDATE emails
		SET status = $1, updated_at = NOW()
		WHERE status = $2 AND updated_at < $3
	`, StatusPending, StatusProcessing, time.Now().Add(-timeout))
	return err
}
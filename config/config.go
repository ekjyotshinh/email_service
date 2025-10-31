package config

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

type Config struct {
	DBHost               string        `json:"db_host"`
	DBUser               string        `json:"db_user"`
	DBPass               string        `json:"db_pass"`
	DBName               string        `json:"db_name"`
	SMTPHost             string        `json:"smtp_host"`
	SMTPPort             int           `json:"smtp_port"`
	SMTPUsername         string        `json:"smtp_username"`
	SMTPPassword         string        `json:"smtp_password"`
	PoolInterval         time.Duration `json:"pool_interval"`
	MaxWorkers           int           `json:"max_workers"`
	MaxRetries           int           `json:"max_retries"`
	BaseBackoffTime      time.Duration `json:"base_backoff_time"`
	ProcessingTimeout    time.Duration `json:"processing_timeout"`
	StuckProcessingCheck time.Duration `json:"stuck_processing_check"`
}

var AppConfig *Config

func LoadConfig() {
	file, err := os.Open("config.json")
	if err != nil {
		log.Fatal("Error opening config file: ", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	AppConfig = &Config{}
	err = decoder.Decode(AppConfig)
	if err != nil {
		log.Fatal("Error decoding config file: ", err)
	}
}
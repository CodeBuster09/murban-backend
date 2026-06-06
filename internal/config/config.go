package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                  string
	GoogleSheetsID        string
	GoogleSheetName       string
	GoogleCredentialsJSON string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Port:                  os.Getenv("PORT"),
		GoogleSheetsID:        os.Getenv("GOOGLE_SHEETS_ID"),
		GoogleSheetName:       os.Getenv("GOOGLE_SHEET_NAME"),
		GoogleCredentialsJSON: os.Getenv("GOOGLE_CREDENTIALS_JSON"),
	}

	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	if cfg.GoogleSheetsID == "" {
		return nil, fmt.Errorf("GOOGLE_SHEETS_ID is not set")
	}
	if cfg.GoogleSheetName == "" {
		cfg.GoogleSheetName = "Sheet1"
	}

	if cfg.GoogleCredentialsJSON == "" {
		credsFile := os.Getenv("GOOGLE_CREDENTIALS_FILE")
		if credsFile == "" {
			return nil, fmt.Errorf("GOOGLE_CREDENTIALS_JSON or GOOGLE_CREDENTIALS_FILE is not set")
		}
		data, err := os.ReadFile(credsFile)
		if err != nil {
			return nil, fmt.Errorf("read GOOGLE_CREDENTIALS_FILE: %w", err)
		}
		cfg.GoogleCredentialsJSON = string(data)
	}

	return cfg, nil
}

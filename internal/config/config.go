package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                  string
	GoogleSheetsID        string
	GoogleSheetName       string
	GoogleCredentialsJSON string
	AllowedOrigins        []string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Port:           os.Getenv("PORT"),
		GoogleSheetsID: os.Getenv("GOOGLE_SHEETS_ID"),
		GoogleSheetName: os.Getenv("GOOGLE_SHEET_NAME"),
	}

	if cfg.Port == "" {
		cfg.Port = "8000"
	}
	if cfg.GoogleSheetsID == "" {
		return nil, fmt.Errorf("GOOGLE_SHEETS_ID is not set")
	}
	if cfg.GoogleSheetName == "" {
		cfg.GoogleSheetName = "Sheet1"
	}

	credsJSON, err := loadGoogleCredentialsJSON()
	if err != nil {
		return nil, err
	}
	cfg.GoogleCredentialsJSON = credsJSON
	cfg.AllowedOrigins = loadAllowedOrigins()

	return cfg, nil
}

func loadAllowedOrigins() []string {
	if raw := strings.TrimSpace(os.Getenv("ALLOWED_ORIGINS")); raw != "" {
		return strings.Split(raw, ",")
	}
	return []string{"https://murban-frontend.onrender.com"}
}

func loadGoogleCredentialsJSON() (string, error) {
	if json := strings.TrimSpace(os.Getenv("GOOGLE_CREDENTIALS_JSON")); json != "" {
		return json, nil
	}

	credsFile := os.Getenv("GOOGLE_CREDENTIALS_FILE")
	if credsFile == "" {
		return "", fmt.Errorf("GOOGLE_CREDENTIALS_JSON or GOOGLE_CREDENTIALS_FILE is not set")
	}
	data, err := os.ReadFile(credsFile)
	if err != nil {
		return "", fmt.Errorf("read GOOGLE_CREDENTIALS_FILE: %w", err)
	}
	return string(data), nil
}

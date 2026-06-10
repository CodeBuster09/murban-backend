package config

import (
	"encoding/base64"
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
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Port:           os.Getenv("PORT"),
		GoogleSheetsID: os.Getenv("GOOGLE_SHEETS_ID"),
		GoogleSheetName: os.Getenv("GOOGLE_SHEET_NAME"),
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

	credsJSON, err := loadGoogleCredentialsJSON()
	if err != nil {
		return nil, err
	}
	cfg.GoogleCredentialsJSON = credsJSON

	return cfg, nil
}

func loadGoogleCredentialsJSON() (string, error) {
	if b64 := strings.TrimSpace(os.Getenv("GOOGLE_CREDENTIALS_JSON_B64")); b64 != "" {
		data, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			return "", fmt.Errorf("decode GOOGLE_CREDENTIALS_JSON_B64: %w", err)
		}
		return string(data), nil
	}

	if json := strings.TrimSpace(os.Getenv("GOOGLE_CREDENTIALS_JSON")); json != "" {
		return json, nil
	}

	credsFile := os.Getenv("GOOGLE_CREDENTIALS_FILE")
	if credsFile == "" {
		return "", fmt.Errorf("GOOGLE_CREDENTIALS_JSON_B64, GOOGLE_CREDENTIALS_JSON, or GOOGLE_CREDENTIALS_FILE is not set")
	}
	data, err := os.ReadFile(credsFile)
	if err != nil {
		return "", fmt.Errorf("read GOOGLE_CREDENTIALS_FILE: %w", err)
	}
	return string(data), nil
}

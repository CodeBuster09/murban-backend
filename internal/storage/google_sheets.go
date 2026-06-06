package storage

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type GoogleSheetsStore struct {
	spreadsheetID string
	sheetName     string
	clientEmail   string
	privateKey    *rsa.PrivateKey
	httpClient    *http.Client

	mu          sync.Mutex
	cachedToken string
	tokenExpiry time.Time
}

func NewGoogleSheetsStore(spreadsheetID, sheetName, credentialsJSON string) (*GoogleSheetsStore, error) {
	var creds struct {
		ClientEmail string `json:"client_email"`
		PrivateKey  string `json:"private_key"`
	}
	if err := json.Unmarshal([]byte(credentialsJSON), &creds); err != nil {
		return nil, fmt.Errorf("parse credentials json: %w", err)
	}
	if creds.ClientEmail == "" || creds.PrivateKey == "" {
		return nil, fmt.Errorf("credentials json missing client_email or private_key")
	}

	block, _ := pem.Decode([]byte(creds.PrivateKey))
	if block == nil {
		return nil, fmt.Errorf("failed to decode private key PEM")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}

	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("private key is not RSA")
	}

	return &GoogleSheetsStore{
		spreadsheetID: spreadsheetID,
		sheetName:     sheetName,
		clientEmail:   creds.ClientEmail,
		privateKey:    rsaKey,
		httpClient:    &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (s *GoogleSheetsStore) AppendFeedback(ctx context.Context, email, message string) error {
	token, err := s.getAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("get access token: %w", err)
	}

	rangeRef := url.PathEscape(s.sheetName) + "!A:C"
	apiURL := fmt.Sprintf(
		"https://sheets.googleapis.com/v4/spreadsheets/%s/values/%s:append?valueInputOption=USER_ENTERED&insertDataOption=INSERT_ROWS",
		s.spreadsheetID,
		rangeRef,
	)

	payload := map[string]any{
		"values": [][]string{{
			time.Now().UTC().Format(time.RFC3339),
			email,
			message,
		}},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("sheets append failed (%d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}

func (s *GoogleSheetsStore) getAccessToken(ctx context.Context) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cachedToken != "" && time.Now().Before(s.tokenExpiry) {
		return s.cachedToken, nil
	}

	now := time.Now()
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`))
	claims := fmt.Sprintf(
		`{"iss":"%s","scope":"https://www.googleapis.com/auth/spreadsheets","aud":"https://oauth2.googleapis.com/token","exp":%d,"iat":%d}`,
		s.clientEmail,
		now.Add(time.Hour).Unix(),
		now.Unix(),
	)
	payload := header + "." + base64.RawURLEncoding.EncodeToString([]byte(claims))

	hash := sha256.Sum256([]byte(payload))
	sig, err := rsa.SignPKCS1v15(rand.Reader, s.privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return "", err
	}

	jwt := payload + "." + base64.RawURLEncoding.EncodeToString(sig)

	form := url.Values{}
	form.Set("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
	form.Set("assertion", jwt)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://oauth2.googleapis.com/token",
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token request failed (%d): %s", resp.StatusCode, string(respBody))
	}

	var tr struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(respBody, &tr); err != nil {
		return "", err
	}

	s.cachedToken = tr.AccessToken
	s.tokenExpiry = time.Now().Add(time.Duration(tr.ExpiresIn-60) * time.Second)

	return s.cachedToken, nil
}

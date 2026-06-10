package main

import (
	"context"
	"log"
	"murban-backend/internal/config"
	"murban-backend/internal/handlers"
	"murban-backend/internal/middleware"
	"murban-backend/internal/storage"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	sheetsStore, err := storage.NewGoogleSheetsStore(
		cfg.GoogleSheetsID,
		cfg.GoogleSheetName,
		cfg.GoogleCredentialsJSON,
	)
	if err != nil {
		log.Fatalf("Failed to init Google Sheets: %v", err)
	}

	feedbackHandler := handlers.NewFeedbackHandler(sheetsStore)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("POST /feedback", feedbackHandler.SubmitFeedback)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      middleware.CORS(cfg.AllowedOrigins, mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("Starting server on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}

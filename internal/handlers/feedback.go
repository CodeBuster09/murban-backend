package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/mail"
	"strings"
)

type FeedbackStore interface {
	AppendFeedback(ctx context.Context, email, message string) error
}

type FeedbackHandler struct {
	store FeedbackStore
}

func NewFeedbackHandler(store FeedbackStore) *FeedbackHandler {
	return &FeedbackHandler{store: store}
}

type feedbackRequest struct {
	Email   string `json:"email"`
	Message string `json:"message"`
}

type feedbackResponse struct {
	Message string `json:"message"`
}

func (h *FeedbackHandler) SubmitFeedback(w http.ResponseWriter, r *http.Request) {
	var req feedbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))
	message := strings.TrimSpace(req.Message)

	if email == "" {
		writeError(w, http.StatusBadRequest, "Email is required")
		return
	}
	if _, err := mail.ParseAddress(email); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid email")
		return
	}

	if err := h.store.AppendFeedback(r.Context(), email, message); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to save feedback")
		return
	}

	writeJSON(w, http.StatusCreated, feedbackResponse{
		Message: "Feedback submitted successfully",
	})
}

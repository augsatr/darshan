package handlers

import (
	"encoding/json"
	"net/http"

	"darshan/api/internal/middleware"
	"darshan/api/internal/models"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) CreateReview(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	slug := chi.URLParam(r, "slug")
	var req models.CreateReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}
	if req.Rating < 1 || req.Rating > 5 {
		http.Error(w, `{"error":"rating must be 1-5"}`, http.StatusBadRequest)
		return
	}

	review, err := h.DB.CreateReview(r.Context(), claims.UserID, slug, req.Rating, req.Text)
	if err != nil {
		http.Error(w, `{"error":"failed to create review"}`, http.StatusInternalServerError)
		return
	}
	if review == nil {
		http.Error(w, `{"error":"temple not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(review)
}

func (h *Handler) ListReviews(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	reviews, err := h.DB.ListReviews(r.Context(), slug)
	if err != nil {
		http.Error(w, `{"error":"failed to list reviews"}`, http.StatusInternalServerError)
		return
	}
	if reviews == nil {
		reviews = []models.Review{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reviews)
}

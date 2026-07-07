package handlers

import (
	"encoding/json"
	"net/http"

	"darshan/api/internal/models"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) ListTemples(w http.ResponseWriter, r *http.Request) {
	if h.DB == nil {
		http.Error(w, `{"error":"database not configured"}`, http.StatusServiceUnavailable)
		return
	}

	temples, err := h.DB.ListTemples(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to fetch temples"}`, http.StatusInternalServerError)
		return
	}

	if temples == nil {
		temples = []models.Temple{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(temples)
}

func (h *Handler) GetTemple(w http.ResponseWriter, r *http.Request) {
	if h.DB == nil {
		http.Error(w, `{"error":"database not configured"}`, http.StatusServiceUnavailable)
		return
	}

	slug := chi.URLParam(r, "slug")
	temple, err := h.DB.GetTempleBySlug(r.Context(), slug)
	if err != nil {
		http.Error(w, `{"error":"failed to fetch temple"}`, http.StatusInternalServerError)
		return
	}
	if temple == nil {
		http.Error(w, `{"error":"temple not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(temple)
}

package handlers

import (
	"encoding/json"
	"net/http"

	"darshan/api/internal/middleware"
	"darshan/api/internal/models"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) AddFavorite(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	slug := chi.URLParam(r, "slug")
	if err := h.DB.AddFavorite(r.Context(), claims.UserID, slug); err != nil {
		http.Error(w, `{"error":"failed to add favorite"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "favorited"})
}

func (h *Handler) RemoveFavorite(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	slug := chi.URLParam(r, "slug")
	if err := h.DB.RemoveFavorite(r.Context(), claims.UserID, slug); err != nil {
		http.Error(w, `{"error":"failed to remove favorite"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "removed"})
}

func (h *Handler) ListFavorites(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	favs, err := h.DB.ListFavorites(r.Context(), claims.UserID)
	if err != nil {
		http.Error(w, `{"error":"failed to list favorites"}`, http.StatusInternalServerError)
		return
	}
	if favs == nil {
		favs = []models.Favorite{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(favs)
}

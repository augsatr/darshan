package handlers

import (
	"encoding/json"
	"net/http"

	"darshan/api/internal/models"

	"github.com/go-chi/chi/v5"
)

const upsertAdminTemple = `INSERT INTO temples (name, slug, state, city, deity, arch_style, description, history,
	image_url, model_url, ambient_audio, latitude, longitude, visit_duration, is_published, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, true, NOW(), NOW())
ON CONFLICT (slug) DO UPDATE SET
	name = EXCLUDED.name, state = EXCLUDED.state, city = EXCLUDED.city,
	deity = EXCLUDED.deity, arch_style = EXCLUDED.arch_style,
	description = EXCLUDED.description, history = EXCLUDED.history,
	image_url = EXCLUDED.image_url, model_url = EXCLUDED.model_url,
	ambient_audio = EXCLUDED.ambient_audio,
	latitude = EXCLUDED.latitude, longitude = EXCLUDED.longitude,
	visit_duration = EXCLUDED.visit_duration,
	is_published = true, updated_at = NOW()`

func (h *Handler) CreateTemple(w http.ResponseWriter, r *http.Request) {
	if h.DB == nil {
		http.Error(w, `{"error":"database not configured"}`, http.StatusServiceUnavailable)
		return
	}

	var t models.Temple
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}
	if t.Name == "" || t.Slug == "" {
		http.Error(w, `{"error":"name and slug required"}`, http.StatusBadRequest)
		return
	}

	_, err := h.DB.Exec(r.Context(), upsertAdminTemple,
		t.Name, t.Slug, t.State, t.City, t.Deity, t.ArchStyle,
		t.Description, t.History, t.ImageURL, t.ModelURL, t.AmbientAudio,
		t.Latitude, t.Longitude, t.VisitDuration,
	)
	if err != nil {
		http.Error(w, `{"error":"failed to save temple"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "saved"})
}

func (h *Handler) DeleteTemple(w http.ResponseWriter, r *http.Request) {
	if h.DB == nil {
		http.Error(w, `{"error":"database not configured"}`, http.StatusServiceUnavailable)
		return
	}

	slug := chi.URLParam(r, "slug")
	_, err := h.DB.Exec(r.Context(), `DELETE FROM temples WHERE slug = $1`, slug)
	if err != nil {
		http.Error(w, `{"error":"failed to delete temple"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

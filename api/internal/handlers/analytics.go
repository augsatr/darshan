package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func (h *Handler) RecordView(w http.ResponseWriter, r *http.Request) {
	if h.DB == nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	templeIDStr := r.URL.Query().Get("temple_id")
	if templeIDStr == "" {
		w.WriteHeader(http.StatusOK)
		return
	}

	templeID, err := strconv.ParseInt(templeIDStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	h.DB.RecordPageView(r.Context(), templeID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "recorded"})
}

func (h *Handler) PopularTemples(w http.ResponseWriter, r *http.Request) {
	if h.DB == nil {
		http.Error(w, `{"error":"database not configured"}`, http.StatusServiceUnavailable)
		return
	}

	limit := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	temples, err := h.DB.PopularTemples(r.Context(), limit)
	if err != nil {
		http.Error(w, `{"error":"failed to fetch popular temples"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(temples)
}

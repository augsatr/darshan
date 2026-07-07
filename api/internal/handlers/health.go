package handlers

import (
	"encoding/json"
	"net/http"
)

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	resp := map[string]string{"status": "ok"}

	if h.DB != nil {
		if err := h.DB.Ping(r.Context()); err != nil {
			resp["database"] = "disconnected"
		} else {
			resp["database"] = "connected"
		}
	} else {
		resp["database"] = "not configured"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

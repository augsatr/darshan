package handlers

import "darshan/api/internal/db"

type Handler struct {
	DB *db.DB
}

func New(d *db.DB) *Handler {
	return &Handler{DB: d}
}

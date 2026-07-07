package handlers

import (
	"darshan/api/internal/auth"
	"darshan/api/internal/db"
)

type Handler struct {
	DB           *db.DB
	RefreshStore *auth.RefreshStore
}

func New(d *db.DB, rs *auth.RefreshStore) *Handler {
	return &Handler{DB: d, RefreshStore: rs}
}

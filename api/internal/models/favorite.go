package models

import "time"

type Favorite struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	TempleID  int64     `json:"temple_id"`
	TempleSlug string   `json:"temple_slug,omitempty"`
	TempleName string   `json:"temple_name,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

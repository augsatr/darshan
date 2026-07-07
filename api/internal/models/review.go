package models

import "time"

type Review struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	TempleID  int64     `json:"temple_id"`
	Rating    int       `json:"rating"`
	Text      string    `json:"text"`
	UserName  string    `json:"user_name,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateReviewRequest struct {
	Rating int    `json:"rating"`
	Text   string `json:"text"`
}

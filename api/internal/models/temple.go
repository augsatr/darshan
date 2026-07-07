package models

import "time"

type Temple struct {
	ID             int64     `json:"id"`
	Name           string    `json:"name"`
	Slug           string    `json:"slug"`
	State          string    `json:"state"`
	City           string    `json:"city"`
	Deity          string    `json:"deity"`
	ArchStyle      string    `json:"arch_style"`
	Description    string    `json:"description"`
	History        string    `json:"history"`
	ImageURL       string    `json:"image_url"`
	ModelURL       string    `json:"model_url"`
	AmbientAudio   string    `json:"ambient_audio"`
	Latitude       float64   `json:"latitude"`
	Longitude      float64   `json:"longitude"`
	VisitDuration  int       `json:"visit_duration"`
	IsPublished    bool      `json:"is_published"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

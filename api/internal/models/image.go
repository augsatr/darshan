package models

type Image struct {
	ID        int64  `json:"id"`
	TempleID  int64  `json:"temple_id"`
	URL       string `json:"url"`
	Alt       string `json:"alt"`
	SortOrder int    `json:"sort_order"`
}

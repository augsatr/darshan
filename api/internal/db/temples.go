package db

import (
	"context"

	"darshan/api/internal/models"

	"github.com/jackc/pgx/v5"
)

const listTemples = `SELECT id, name, slug, state, city, deity, arch_style, description, history,
	image_url, model_url, ambient_audio, latitude, longitude, visit_duration,
	is_published, created_at, updated_at
FROM temples WHERE is_published = true ORDER BY name`

const getTempleBySlug = `SELECT id, name, slug, state, city, deity, arch_style, description, history,
	image_url, model_url, ambient_audio, latitude, longitude, visit_duration,
	is_published, created_at, updated_at
FROM temples WHERE slug = $1`

func scanTemple(row pgx.Row) (models.Temple, error) {
	var t models.Temple
	err := row.Scan(
		&t.ID, &t.Name, &t.Slug, &t.State, &t.City, &t.Deity,
		&t.ArchStyle, &t.Description, &t.History, &t.ImageURL,
		&t.ModelURL, &t.AmbientAudio, &t.Latitude, &t.Longitude,
		&t.VisitDuration, &t.IsPublished, &t.CreatedAt, &t.UpdatedAt,
	)
	return t, err
}

func (d *DB) ListTemples(ctx context.Context) ([]models.Temple, error) {
	rows, err := d.Query(ctx, listTemples)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var temples []models.Temple
	for rows.Next() {
		t, err := scanTemple(rows)
		if err != nil {
			return nil, err
		}
		temples = append(temples, t)
	}
	return temples, rows.Err()
}

func (d *DB) GetTempleBySlug(ctx context.Context, slug string) (*models.Temple, error) {
	t, err := scanTemple(d.QueryRow(ctx, getTempleBySlug, slug))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

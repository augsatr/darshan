package db

import (
	"context"

	"darshan/api/internal/models"

	"github.com/jackc/pgx/v5"
)

const addFavorite = `INSERT INTO favorites (user_id, temple_id) VALUES ($1, $2)
ON CONFLICT DO NOTHING RETURNING id, created_at`

const removeFavorite = `DELETE FROM favorites WHERE user_id = $1 AND temple_id = $2`

const listFavorites = `SELECT f.id, f.user_id, f.temple_id, t.slug, t.name, f.created_at
FROM favorites f JOIN temples t ON t.id = f.temple_id
WHERE f.user_id = $1 ORDER BY f.created_at DESC`

const getTempleIDBySlug = `SELECT id FROM temples WHERE slug = $1`

func (d *DB) AddFavorite(ctx context.Context, userID int64, templeSlug string) error {
	var templeID int64
	err := d.QueryRow(ctx, getTempleIDBySlug, templeSlug).Scan(&templeID)
	if err != nil {
		return err
	}
	_, err = d.Exec(ctx, addFavorite, userID, templeID)
	return err
}

func (d *DB) RemoveFavorite(ctx context.Context, userID int64, templeSlug string) error {
	var templeID int64
	err := d.QueryRow(ctx, getTempleIDBySlug, templeSlug).Scan(&templeID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil
		}
		return err
	}
	_, err = d.Exec(ctx, removeFavorite, userID, templeID)
	return err
}

func (d *DB) ListFavorites(ctx context.Context, userID int64) ([]models.Favorite, error) {
	rows, err := d.Query(ctx, listFavorites, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var favs []models.Favorite
	for rows.Next() {
		var f models.Favorite
		if err := rows.Scan(&f.ID, &f.UserID, &f.TempleID, &f.TempleSlug, &f.TempleName, &f.CreatedAt); err != nil {
			return nil, err
		}
		favs = append(favs, f)
	}
	return favs, rows.Err()
}

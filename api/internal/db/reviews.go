package db

import (
	"context"

	"darshan/api/internal/models"

	"github.com/jackc/pgx/v5"
)

const createReview = `INSERT INTO reviews (user_id, temple_id, rating, text) VALUES ($1, $2, $3, $4)
RETURNING id, created_at, updated_at`

const listReviews = `SELECT r.id, r.user_id, r.temple_id, r.rating, r.text, u.name, r.created_at, r.updated_at
FROM reviews r JOIN users u ON u.id = r.user_id
WHERE r.temple_id = $1 ORDER BY r.created_at DESC`

const getTempleID = `SELECT id FROM temples WHERE slug = $1`

func (d *DB) CreateReview(ctx context.Context, userID int64, slug string, rating int, text string) (*models.Review, error) {
	var templeID int64
	err := d.QueryRow(ctx, getTempleID, slug).Scan(&templeID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	var r models.Review
	err = d.QueryRow(ctx, createReview, userID, templeID, rating, text).Scan(&r.ID, &r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		return nil, err
	}
	r.UserID = userID
	r.TempleID = templeID
	r.Rating = rating
	r.Text = text
	return &r, nil
}

func (d *DB) ListReviews(ctx context.Context, slug string) ([]models.Review, error) {
	var templeID int64
	err := d.QueryRow(ctx, getTempleID, slug).Scan(&templeID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	rows, err := d.Query(ctx, listReviews, templeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []models.Review
	for rows.Next() {
		var r models.Review
		if err := rows.Scan(&r.ID, &r.UserID, &r.TempleID, &r.Rating, &r.Text, &r.UserName, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		reviews = append(reviews, r)
	}
	return reviews, rows.Err()
}

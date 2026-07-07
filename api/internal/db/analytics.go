package db

import (
	"context"
)

const recordView = `INSERT INTO page_views (temple_id) VALUES ($1)`

const popularTemples = `SELECT t.id, t.name, t.slug, t.state, COUNT(pv.id) as views
FROM page_views pv JOIN temples t ON t.id = pv.temple_id
GROUP BY t.id ORDER BY views DESC LIMIT $1`

type PopularTemple struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Slug  string `json:"slug"`
	State string `json:"state"`
	Views int    `json:"views"`
}

func (d *DB) RecordPageView(ctx context.Context, templeID int64) error {
	_, err := d.Exec(ctx, recordView, templeID)
	return err
}

func (d *DB) PopularTemples(ctx context.Context, limit int) ([]PopularTemple, error) {
	rows, err := d.Query(ctx, popularTemples, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []PopularTemple
	for rows.Next() {
		var p PopularTemple
		if err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.State, &p.Views); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, rows.Err()
}

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"darshan/api/internal/db"
	"darshan/api/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

const upsertSQL = `INSERT INTO temples (name, slug, state, city, deity, arch_style, description, history,
	image_url, model_url, ambient_audio, latitude, longitude, visit_duration, is_published, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, true, NOW(), NOW())
ON CONFLICT (slug) DO UPDATE SET
	name = EXCLUDED.name, state = EXCLUDED.state, city = EXCLUDED.city,
	deity = EXCLUDED.deity, arch_style = EXCLUDED.arch_style,
	description = EXCLUDED.description, history = EXCLUDED.history,
	image_url = EXCLUDED.image_url, model_url = EXCLUDED.model_url,
	ambient_audio = EXCLUDED.ambient_audio,
	latitude = EXCLUDED.latitude, longitude = EXCLUDED.longitude,
	visit_duration = EXCLUDED.visit_duration,
	is_published = true, updated_at = NOW()`

func main() {
	godotenv.Load()

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		slog.Error("DATABASE_URL is required")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := db.Connect(ctx, connStr)
	if err != nil {
		slog.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	jsonFile := "seed/temples.json"
	if len(os.Args) > 1 {
		jsonFile = os.Args[1]
	}

	data, err := os.ReadFile(jsonFile)
	if err != nil {
		slog.Error("failed to read seed file", "path", jsonFile, "error", err)
		os.Exit(1)
	}

	var temples []models.Temple
	if err := json.Unmarshal(data, &temples); err != nil {
		slog.Error("failed to parse JSON", "error", err)
		os.Exit(1)
	}

	now := time.Now()
	batch := &pgx.Batch{}
	for _, t := range temples {
		batch.Queue(upsertSQL,
			t.Name, t.Slug, t.State, t.City, t.Deity, t.ArchStyle,
			t.Description, t.History, t.ImageURL, t.ModelURL, t.AmbientAudio,
			t.Latitude, t.Longitude, t.VisitDuration,
		)
	}

	br := pool.SendBatch(ctx, batch)
	defer br.Close()

	var count int
	for range temples {
		_, err := br.Exec()
		if err != nil {
			slog.Error("seed failed", "error", err)
			os.Exit(1)
		}
		count++
	}

	fmt.Printf("Seeded %d temples (%.0f ms)\n", count, float64(time.Since(now).Microseconds())/1000)
}

package db

import (
	"context"

	"darshan/api/internal/models"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

const createUser = `INSERT INTO users (email, password_hash, name) VALUES ($1, $2, $3)
RETURNING id, email, name, role, created_at`

const getUserByEmail = `SELECT id, email, password_hash, name, role, created_at FROM users WHERE email = $1`

const getUserByID = `SELECT id, email, name, role, created_at FROM users WHERE id = $1`

func (d *DB) CreateUser(ctx context.Context, req models.SignupRequest) (*models.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	var u models.User
	err = d.QueryRow(ctx, createUser, req.Email, string(hash), req.Name).Scan(
		&u.ID, &u.Email, &u.Name, &u.Role, &u.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (d *DB) Authenticate(ctx context.Context, email, password string) (*models.User, error) {
	var u models.User
	var hash string
	err := d.QueryRow(ctx, getUserByEmail, email).Scan(&u.ID, &u.Email, &hash, &u.Name, &u.Role, &u.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return nil, nil
	}
	return &u, nil
}

func (d *DB) GetUser(ctx context.Context, id int64) (*models.User, error) {
	var u models.User
	err := d.QueryRow(ctx, getUserByID, id).Scan(&u.ID, &u.Email, &u.Name, &u.Role, &u.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

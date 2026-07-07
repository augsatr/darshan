package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const refreshTokenTTL = 30 * 24 * time.Hour

var (
	ErrTokenNotFound = errors.New("refresh token not found")
	ErrTokenExpired  = errors.New("refresh token expired")
	ErrTokenRevoked  = errors.New("refresh token revoked")
	ErrTokenReused   = errors.New("refresh token reuse detected")
)

type RefreshStore struct {
	db *pgxpool.Pool
}

func NewRefreshStore(db *pgxpool.Pool) *RefreshStore {
	return &RefreshStore{db: db}
}

func generateToken() (raw string, hash string, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	raw = base64.RawURLEncoding.EncodeToString(b)
	h := sha256.Sum256([]byte(raw))
	hash = hex.EncodeToString(h[:])
	return raw, hash, nil
}

func (s *RefreshStore) Issue(ctx context.Context, userID int64, userAgent, ip string) (string, error) {
	raw, hash, err := generateToken()
	if err != nil {
		return "", err
	}

	_, err = s.db.Exec(ctx, `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at, user_agent, ip_address)
		VALUES ($1, $2, $3, $4, $5)
	`, userID, hash, time.Now().Add(refreshTokenTTL), userAgent, ip)
	if err != nil {
		return "", err
	}

	return raw, nil
}

func (s *RefreshStore) Rotate(ctx context.Context, rawToken, userAgent, ip string) (newRaw string, userID int64, err error) {
	h := sha256.Sum256([]byte(rawToken))
	hash := hex.EncodeToString(h[:])

	var (
		id        int64
		uid       int64
		expiresAt time.Time
		revokedAt *time.Time
	)

	row := s.db.QueryRow(ctx, `
		SELECT id, user_id, expires_at, revoked_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`, hash)

	if err := row.Scan(&id, &uid, &expiresAt, &revokedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", 0, ErrTokenNotFound
		}
		return "", 0, err
	}

	if revokedAt != nil {
		_ = s.revokeAllForUser(ctx, uid)
		return "", 0, ErrTokenReused
	}

	if time.Now().After(expiresAt) {
		return "", 0, ErrTokenExpired
	}

	newTokenRaw, newHash, err := generateToken()
	if err != nil {
		return "", 0, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return "", 0, err
	}
	defer tx.Rollback(ctx)

	var newID int64
	err = tx.QueryRow(ctx, `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at, user_agent, ip_address)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, uid, newHash, time.Now().Add(refreshTokenTTL), userAgent, ip).Scan(&newID)
	if err != nil {
		return "", 0, err
	}

	_, err = tx.Exec(ctx, `
		UPDATE refresh_tokens
		SET revoked_at = now(), replaced_by = $1
		WHERE id = $2
	`, newID, id)
	if err != nil {
		return "", 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return "", 0, err
	}

	return newTokenRaw, uid, nil
}

func (s *RefreshStore) Revoke(ctx context.Context, rawToken string) error {
	h := sha256.Sum256([]byte(rawToken))
	hash := hex.EncodeToString(h[:])

	_, err := s.db.Exec(ctx, `
		UPDATE refresh_tokens SET revoked_at = now()
		WHERE token_hash = $1 AND revoked_at IS NULL
	`, hash)
	return err
}

func (s *RefreshStore) revokeAllForUser(ctx context.Context, userID int64) error {
	_, err := s.db.Exec(ctx, `
		UPDATE refresh_tokens SET revoked_at = now()
		WHERE user_id = $1 AND revoked_at IS NULL
	`, userID)
	return err
}

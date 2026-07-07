package handlers

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"darshan/api/internal/auth"
	"darshan/api/internal/db"
	authmw "darshan/api/internal/middleware"
)

// newTestRouter wires a fresh Handler against testPool for a single test.
//
// ASSUMPTION (unconfirmed): db.DB{Pool: testPool} is a valid zero-value-
// otherwise construction — only holds if db.DB has no other unexported state
// needing init. Flag if that's wrong; a db.NewForTest(pool) would be cleaner.
func newTestRouter(t *testing.T) (*chi.Mux, *Handler) {
	t.Helper()
	truncateAll(t)

	// Reset the package-level rate limiter so state doesn't leak between tests.
	loginLimiter = authmw.NewLoginRateLimiter(5, 5*time.Minute, 15*time.Minute)

	d := &db.DB{Pool: testPool}
	refreshStore := auth.NewRefreshStore(testPool)
	h := New(d, refreshStore)

	r := chi.NewRouter()
	r.Post("/auth/signup", h.Signup)
	r.Post("/auth/login", h.Login)
	r.Post("/auth/refresh", h.RefreshToken)
	r.Post("/auth/logout", h.Logout)
	// Auth middleware populates claims via GetClaims — without it, Me sees
	// nil claims regardless of the bearer token supplied.
	r.With(authmw.Auth).Get("/auth/me", h.Me)

	return r, h
}

func signupTestUser(t *testing.T, r *chi.Mux, email, password string) *http.Cookie {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"email": email, "password": password})
	req := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated && rec.Code != http.StatusOK {
		t.Fatalf("signup failed: status %d, body %s", rec.Code, rec.Body.String())
	}
	return findCookie(rec.Result().Cookies(), "refresh_token")
}

func findCookie(cookies []*http.Cookie, name string) *http.Cookie {
	for _, c := range cookies {
		if c.Name == name {
			return c
		}
	}
	return nil
}

// hashRefreshToken mirrors the SHA-256 hex hashing used in internal/auth to
// store refresh tokens. ASSUMPTION (unconfirmed): plain sha256 hex, no salt,
// column named token_hash. Still not verified against refresh.go — if wrong,
// TestRefresh_ExpiredToken's UPDATE will match zero rows.
func hashRefreshToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// ---- Login ----

func TestLogin_ValidCredentials_ReturnsTokens(t *testing.T) {
	r, _ := newTestRouter(t)
	signupTestUser(t, r, "ray@example.com", "correct-horse-battery-staple")

	body, _ := json.Marshal(map[string]string{"email": "ray@example.com", "password": "correct-horse-battery-staple"})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	// Field is "token" per models.AuthResponse, not "access_token".
	var resp struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Token == "" {
		t.Fatal("expected non-empty token")
	}
	if findCookie(rec.Result().Cookies(), "refresh_token") == nil {
		t.Fatal("expected refresh_token cookie to be set")
	}
}

func TestLogin_InvalidCredentials_Returns401AndRecordsFailure(t *testing.T) {
	r, _ := newTestRouter(t)
	signupTestUser(t, r, "ray@example.com", "correct-horse-battery-staple")

	body, _ := json.Marshal(map[string]string{"email": "ray@example.com", "password": "wrong-password"})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestLogin_RateLimitTrips_Returns429(t *testing.T) {
	r, _ := newTestRouter(t)
	signupTestUser(t, r, "ray@example.com", "correct-horse-battery-staple")

	body, _ := json.Marshal(map[string]string{"email": "ray@example.com", "password": "wrong-password"})

	var lastCode int
	for i := 0; i < 6; i++ {
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.RemoteAddr = "203.0.113.5:1234"
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		lastCode = rec.Code
	}

	if lastCode != http.StatusTooManyRequests {
		t.Fatalf("expected 429 on 6th attempt, got %d", lastCode)
	}
}

// ---- Refresh ----

func TestRefresh_ValidToken_ReturnsNewPairAndRevokesOld(t *testing.T) {
	r, _ := newTestRouter(t)
	cookie := signupTestUser(t, r, "ray@example.com", "correct-horse-battery-staple")
	if cookie == nil {
		t.Fatal("expected refresh cookie from signup")
	}

	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	newCookie := findCookie(rec.Result().Cookies(), "refresh_token")
	if newCookie == nil || newCookie.Value == cookie.Value {
		t.Fatal("expected a rotated refresh token, different from the original")
	}

	req2 := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	req2.AddCookie(cookie)
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 reusing rotated-out token, got %d", rec2.Code)
	}
}

func TestRefresh_ReusedToken_KillsAllSessions(t *testing.T) {
	r, _ := newTestRouter(t)
	cookie := signupTestUser(t, r, "ray@example.com", "correct-horse-battery-staple")

	req1 := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	req1.AddCookie(cookie)
	rec1 := httptest.NewRecorder()
	r.ServeHTTP(rec1, req1)
	newCookie := findCookie(rec1.Result().Cookies(), "refresh_token")
	if rec1.Code != http.StatusOK || newCookie == nil {
		t.Fatalf("setup refresh failed: %d %s", rec1.Code, rec1.Body.String())
	}

	req2 := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	req2.AddCookie(cookie)
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 on reused token, got %d", rec2.Code)
	}

	req3 := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	req3.AddCookie(newCookie)
	rec3 := httptest.NewRecorder()
	r.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusUnauthorized {
		t.Fatalf("expected reuse detection to also kill the legitimate rotated token, got %d", rec3.Code)
	}
}

func TestRefresh_ExpiredToken_Returns401(t *testing.T) {
	r, _ := newTestRouter(t)
	cookie := signupTestUser(t, r, "ray@example.com", "correct-horse-battery-staple")

	hash := hashRefreshToken(cookie.Value)
	_, err := testPool.Exec(context.Background(),
		`UPDATE refresh_tokens SET expires_at = now() - interval '1 minute' WHERE token_hash = $1`, hash)
	if err != nil {
		t.Fatalf("failed to force-expire token: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for expired token, got %d", rec.Code)
	}
}

// ---- Logout ----

func TestLogout_RevokesToken_SubsequentRefreshFails(t *testing.T) {
	r, _ := newTestRouter(t)
	cookie := signupTestUser(t, r, "ray@example.com", "correct-horse-battery-staple")

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK && rec.Code != http.StatusNoContent {
		t.Fatalf("expected 200/204 on logout, got %d: %s", rec.Code, rec.Body.String())
	}

	req2 := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	req2.AddCookie(cookie)
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 refreshing after logout, got %d", rec2.Code)
	}
}

// ---- /auth/me ----

func TestMe_ValidAccessToken_Returns200(t *testing.T) {
	r, _ := newTestRouter(t)
	signupTestUser(t, r, "ray@example.com", "correct-horse-battery-staple")

	body, _ := json.Marshal(map[string]string{"email": "ray@example.com", "password": "correct-horse-battery-staple"})
	loginReq := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	r.ServeHTTP(loginRec, loginReq)

	var loginResp struct {
		Token string `json:"token"`
	}
	json.NewDecoder(loginRec.Body).Decode(&loginResp)

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+loginResp.Token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestMe_MissingOrExpiredToken_Returns401(t *testing.T) {
	r, _ := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 with no token, got %d", rec.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req2.Header.Set("Authorization", "Bearer not-a-real-token")
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 with garbage token, got %d", rec2.Code)
	}
}

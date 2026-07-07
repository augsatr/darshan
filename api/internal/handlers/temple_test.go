package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	authmw "darshan/api/internal/middleware"
)

// newTestRouterWithTemples wires public read routes plus admin write routes.
//
// ASSUMPTIONS still unconfirmed:
//   - Public GetTemple is keyed by slug, same as admin routes (only admin
//     routes were explicitly confirmed as slug-based) — flag if List/Get are
//     actually id-based instead.
//   - Admin write routes need both authmw.Auth and authmw.Admin chained (vs.
//     Admin doing its own auth check internally) — chained here to be safe;
//     drop authmw.Auth from the chain if Admin already covers it.
func newTestRouterWithTemples(t *testing.T) (*chi.Mux, *Handler) {
	t.Helper()
	r, h := newTestRouter(t)
	r.Get("/temples", h.ListTemples)
	r.Get("/temples/{slug}", h.GetTemple)
	r.With(authmw.Auth, authmw.Admin).Post("/admin/temples", h.CreateTemple)
	r.With(authmw.Auth, authmw.Admin).Delete("/admin/temples/{slug}", h.DeleteTemple)
	return r, h
}

func accessTokenFor(t *testing.T, r *chi.Mux, email string, admin bool) string {
	t.Helper()
	signupTestUser(t, r, email, "correct-horse-battery-staple")

	if admin {
		_, err := testPool.Exec(context.Background(), `UPDATE users SET role = 'admin' WHERE email = $1`, email)
		if err != nil {
			t.Fatalf("promote to admin: %v", err)
		}
	}

	body, _ := json.Marshal(map[string]string{"email": email, "password": "correct-horse-battery-staple"})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	var resp struct {
		Token string `json:"token"`
	}
	json.NewDecoder(rec.Body).Decode(&resp)
	return resp.Token
}

// seedTemple inserts a temple directly and returns its (id, slug).
//
// ASSUMPTION (unconfirmed): "Maharashtra" is a placeholder for the required
// `state` column. If there's a CHECK constraint restricting valid values,
// swap this for whatever your seed data (cmd/seed/main.go) actually uses.
func seedTemple(t *testing.T, name, slug string) (int64, string) {
	t.Helper()
	var id int64
	err := testPool.QueryRow(context.Background(),
		`INSERT INTO temples (name, slug, state) VALUES ($1, $2, $3) RETURNING id`,
		name, slug, "Maharashtra",
	).Scan(&id)
	if err != nil {
		t.Fatalf("seed temple: %v", err)
	}
	return id, slug
}

// ---- Read (public) ----

func TestListTemples_NoAuthRequired(t *testing.T) {
	r, _ := newTestRouterWithTemples(t)
	seedTemple(t, "Somnath Temple", "somnath-temple")

	req := httptest.NewRequest(http.MethodGet, "/temples", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGetTemple_NoAuthRequired(t *testing.T) {
	r, _ := newTestRouterWithTemples(t)
	_, slug := seedTemple(t, "Somnath Temple", "somnath-temple")

	req := httptest.NewRequest(http.MethodGet, "/temples/"+slug, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

// ---- Create ----

func TestCreateTemple_NonAdmin_Returns403(t *testing.T) {
	r, _ := newTestRouterWithTemples(t)
	token := accessTokenFor(t, r, "regular@example.com", false)

	body, _ := json.Marshal(map[string]string{"name": "New Temple", "slug": "new-temple", "state": "Maharashtra"})
	req := httptest.NewRequest(http.MethodPost, "/admin/temples", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for non-admin create, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateTemple_Admin_ReturnsSavedAndPersists(t *testing.T) {
	r, _ := newTestRouterWithTemples(t)
	token := accessTokenFor(t, r, "admin@example.com", true)

	body, _ := json.Marshal(map[string]string{"name": "New Temple", "slug": "new-temple", "state": "Maharashtra"})
	req := httptest.NewRequest(http.MethodPost, "/admin/temples", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// CreateTemple returns 200 with {"status":"saved"}, not 201.
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for admin create, got %d: %s", rec.Code, rec.Body.String())
	}

	var row struct {
		Name string
		Slug string
	}
	err := testPool.QueryRow(context.Background(), `SELECT name, slug FROM temples WHERE slug = $1`, "new-temple").
		Scan(&row.Name, &row.Slug)
	if err != nil {
		t.Fatalf("expected temple to be persisted: %v", err)
	}
	if row.Name != "New Temple" {
		t.Fatalf("expected persisted name %q, got %q", "New Temple", row.Name)
	}
}

// ---- Delete ----
// (No UpdateTemple handler exists in the codebase — update tests removed.)

func TestDeleteTemple_NonAdmin_Returns403(t *testing.T) {
	r, _ := newTestRouterWithTemples(t)
	_, slug := seedTemple(t, "Somnath Temple", "somnath-temple")
	token := accessTokenFor(t, r, "regular@example.com", false)

	req := httptest.NewRequest(http.MethodDelete, "/admin/temples/"+slug, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestDeleteTemple_Admin_RemovesRow(t *testing.T) {
	r, _ := newTestRouterWithTemples(t)
	id, slug := seedTemple(t, "Somnath Temple", "somnath-temple")
	token := accessTokenFor(t, r, "admin@example.com", true)

	req := httptest.NewRequest(http.MethodDelete, "/admin/temples/"+slug, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// DeleteTemple returns 200 with {"status":"deleted"}, not 204.
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var count int
	testPool.QueryRow(context.Background(), `SELECT count(*) FROM temples WHERE id = $1`, id).Scan(&count)
	if count != 0 {
		t.Fatalf("expected temple to be removed, still found %d rows", count)
	}
}

// ---- Input validation ----

func TestCreateTemple_MalformedBody_Returns400NotServerError(t *testing.T) {
	r, _ := newTestRouterWithTemples(t)
	token := accessTokenFor(t, r, "admin@example.com", true)

	req := httptest.NewRequest(http.MethodPost, "/admin/temples", bytes.NewReader([]byte(`{"name": `)))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for malformed body, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateTemple_MissingRequiredFields_Returns400(t *testing.T) {
	r, _ := newTestRouterWithTemples(t)
	token := accessTokenFor(t, r, "admin@example.com", true)

	body, _ := json.Marshal(map[string]string{}) // no name/slug/state
	req := httptest.NewRequest(http.MethodPost, "/admin/temples", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing required fields, got %d: %s", rec.Code, rec.Body.String())
	}
}

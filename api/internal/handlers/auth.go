package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"darshan/api/internal/auth"
	"darshan/api/internal/middleware"
	"darshan/api/internal/models"
)

var loginLimiter = middleware.NewLoginRateLimiter(5, 5*time.Minute, 15*time.Minute)

func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) {
	if h.DB == nil {
		http.Error(w, `{"error":"database not configured"}`, http.StatusServiceUnavailable)
		return
	}

	var req models.SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}
	if req.Email == "" || req.Password == "" {
		http.Error(w, `{"error":"email and password required"}`, http.StatusBadRequest)
		return
	}

	user, err := h.DB.CreateUser(r.Context(), req)
	if err != nil {
		http.Error(w, `{"error":"email already registered"}`, http.StatusConflict)
		return
	}

	accessToken, err := auth.GenerateAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		http.Error(w, `{"error":"failed to generate token"}`, http.StatusInternalServerError)
		return
	}

	if h.RefreshStore != nil {
		refreshToken, err := h.RefreshStore.Issue(r.Context(), user.ID, r.UserAgent(), middleware.ClientIP(r))
		if err == nil {
			setRefreshCookie(w, r, refreshToken)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.AuthResponse{Token: accessToken, User: *user})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if h.DB == nil {
		http.Error(w, `{"error":"database not configured"}`, http.StatusServiceUnavailable)
		return
	}

	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	ip := middleware.ClientIP(r)

	if !loginLimiter.Allow(ip, req.Email) {
		http.Error(w, `{"error":"too many login attempts, try again later"}`, http.StatusTooManyRequests)
		return
	}

	user, err := h.DB.Authenticate(r.Context(), req.Email, req.Password)
	if err != nil {
		http.Error(w, `{"error":"server error"}`, http.StatusInternalServerError)
		return
	}
	if user == nil {
		loginLimiter.RecordFailure(ip, req.Email)
		http.Error(w, `{"error":"invalid email or password"}`, http.StatusUnauthorized)
		return
	}

	loginLimiter.Reset(ip, req.Email)

	accessToken, err := auth.GenerateAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		http.Error(w, `{"error":"failed to generate token"}`, http.StatusInternalServerError)
		return
	}

	var refreshToken string
	if h.RefreshStore != nil {
		refreshToken, err = h.RefreshStore.Issue(r.Context(), user.ID, r.UserAgent(), ip)
		if err != nil {
			http.Error(w, `{"error":"failed to issue refresh token"}`, http.StatusInternalServerError)
			return
		}
		setRefreshCookie(w, r, refreshToken)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.AuthResponse{Token: accessToken, User: *user})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	user, err := h.DB.GetUser(r.Context(), claims.UserID)
	if err != nil || user == nil {
		http.Error(w, `{"error":"user not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	if h.RefreshStore == nil {
		http.Error(w, `{"error":"not available"}`, http.StatusServiceUnavailable)
		return
	}

	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		http.Error(w, `{"error":"no refresh token"}`, http.StatusUnauthorized)
		return
	}

	newRaw, userID, err := h.RefreshStore.Rotate(r.Context(), cookie.Value, r.UserAgent(), middleware.ClientIP(r))
	if err != nil {
		if errors.Is(err, auth.ErrTokenReused) {
			clearRefreshCookie(w, r)
			http.Error(w, `{"error":"session compromised, please log in again"}`, http.StatusUnauthorized)
			return
		}
		http.Error(w, `{"error":"invalid session"}`, http.StatusUnauthorized)
		return
	}

	user, err := h.DB.GetUser(r.Context(), userID)
	if err != nil || user == nil {
		http.Error(w, `{"error":"user not found"}`, http.StatusUnauthorized)
		return
	}

	accessToken, err := auth.GenerateAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		http.Error(w, `{"error":"failed to generate token"}`, http.StatusInternalServerError)
		return
	}

	setRefreshCookie(w, r, newRaw)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"access_token": accessToken})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if h.RefreshStore != nil {
		cookie, err := r.Cookie("refresh_token")
		if err == nil {
			_ = h.RefreshStore.Revoke(r.Context(), cookie.Value)
		}
	}
	clearRefreshCookie(w, r)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "logged out"})
}

func setRefreshCookie(w http.ResponseWriter, r *http.Request, token string) {
	secure := r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteNoneMode,
		MaxAge:   int((30 * 24 * time.Hour).Seconds()),
	})
}

func clearRefreshCookie(w http.ResponseWriter, r *http.Request) {
	secure := r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteNoneMode,
		MaxAge:   -1,
	})
}

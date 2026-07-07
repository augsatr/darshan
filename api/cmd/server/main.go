package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"darshan/api/internal/auth"
	"darshan/api/internal/db"
	"darshan/api/internal/handlers"
	authmw "darshan/api/internal/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func cacheControl(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=60, s-maxage=120")
		next.ServeHTTP(w, r)
	})
}

func main() {
	godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		slog.Warn("DATABASE_URL not set, running without database")
	}

	var d *db.DB
	var err error
	if connStr != "" {
		d, err = db.Connect(context.Background(), connStr)
		if err != nil {
			slog.Error("failed to connect to database", "error", err)
			os.Exit(1)
		}
		defer d.Close()
	}

	r := chi.NewRouter()
	r.Use(cors)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.Timeout(30 * time.Second))

	var rs *auth.RefreshStore
	if d != nil {
		rs = auth.NewRefreshStore(d.Pool)
	}
	h := handlers.New(d, rs)

	r.Get("/health", h.Health)
	r.With(cacheControl).Get("/temples", h.ListTemples)
	r.With(cacheControl).Get("/temples/{slug}", h.GetTemple)

	r.Post("/auth/signup", h.Signup)
	r.Post("/auth/login", h.Login)
	r.Post("/auth/refresh", h.RefreshToken)
	r.Post("/auth/logout", h.Logout)

	r.With(authmw.Auth).Get("/auth/me", h.Me)
	r.With(authmw.Auth).Get("/favorites", h.ListFavorites)
	r.With(authmw.Auth).Post("/favorites/{slug}", h.AddFavorite)
	r.With(authmw.Auth).Delete("/favorites/{slug}", h.RemoveFavorite)
	r.With(authmw.Auth).Post("/reviews/{slug}", h.CreateReview)

	r.Get("/reviews/{slug}", h.ListReviews)
	r.Post("/views", h.RecordView)
	r.Get("/popular", h.PopularTemples)

	r.With(authmw.Admin).Post("/admin/temples", h.CreateTemple)
	r.With(authmw.Admin).Delete("/admin/temples/{slug}", h.DeleteTemple)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("server starting", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("forced shutdown", "error", err)
		os.Exit(1)
	}
	slog.Info("server stopped")
}

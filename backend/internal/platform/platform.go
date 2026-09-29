package platform

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func Env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func OpenDB() (*sql.DB, error) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		return nil, errors.New("DATABASE_URL is required")
	}
	db, err := sql.Open("pgx", url)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(12)
	db.SetMaxIdleConns(4)
	db.SetConnMaxLifetime(30 * time.Minute)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func Serve(addr string, handler http.Handler) error {
	server := &http.Server{Addr: addr, Handler: handler, ReadHeaderTimeout: 5 * time.Second}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		deadline, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(deadline); err != nil {
			slog.Error("shutdown failed", "error", err)
		}
	}()
	slog.Info("listening", "addr", addr)
	err := server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func Health(db *sql.DB) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health/live", func(w http.ResponseWriter, r *http.Request) { JSON(w, 200, map[string]string{"status": "ok"}) })
	mux.HandleFunc("GET /health/ready", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if db != nil && db.PingContext(ctx) != nil {
			JSON(w, 503, map[string]string{"status": "unready"})
			return
		}
		JSON(w, 200, map[string]string{"status": "ready"})
	})
	return mux
}

func JSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.Error("encode response", "error", err)
	}
}

func Error(w http.ResponseWriter, status int, code, message string) {
	JSON(w, status, map[string]any{"error": map[string]any{"code": code, "message": message, "retryable": status >= 500, "requestId": ""}})
}

func Decode(r *http.Request, value any) error {
	data, err := io.ReadAll(io.LimitReader(r.Body, 64*1024+1))
	if err != nil {
		return err
	}
	if len(data) > 64*1024 {
		return errors.New("request body too large")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return errors.New("more than one JSON value")
		}
		return err
	}
	return nil
}

func Internal(next http.Handler) http.Handler {
	expected := os.Getenv("INTERNAL_TOKEN")
	if len(expected) < 24 {
		panic("INTERNAL_TOKEN must be at least 24 characters")
	}
	if expected == "local-internal-token-change-in-production" && os.Getenv("APP_ENV") != "local" {
		panic("development INTERNAL_TOKEN is forbidden outside local mode")
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health/live" || r.URL.Path == "/health/ready" {
			next.ServeHTTP(w, r)
			return
		}
		if !EqualToken(r.Header.Get("X-Internal-Token"), expected) {
			Error(w, 401, "unauthorized", "Internal authentication required")
			return
		}
		if r.Header.Get("X-User-ID") == "" {
			Error(w, 400, "missing_user", "User identity required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func EqualToken(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var mismatch byte
	for i := range a {
		mismatch |= a[i] ^ b[i]
	}
	return mismatch == 0
}

func User(r *http.Request) string { return r.Header.Get("X-User-ID") }

func CheckUUID(id string) error {
	if len(id) != 36 {
		return fmt.Errorf("invalid id")
	}
	for i, c := range id {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if c != '-' {
				return fmt.Errorf("invalid id")
			}
			continue
		}
		if !strings.ContainsRune("0123456789abcdefABCDEF", c) {
			return fmt.Errorf("invalid id")
		}
	}
	return nil
}

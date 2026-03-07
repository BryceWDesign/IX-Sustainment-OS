package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	defaultAddr         = ":8080"
	defaultReadTimeout  = 10 * time.Second
	defaultWriteTimeout = 15 * time.Second
	defaultIdleTimeout  = 60 * time.Second
	shutdownTimeout     = 10 * time.Second
)

type appInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Environment string `json:"environment"`
	Status      string `json:"status"`
	TimeUTC     string `json:"time_utc"`
}

type healthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	TimeUTC string `json:"time_utc"`
}

func main() {
	logger := log.New(os.Stdout, "[ix-sustainment-os] ", log.LstdFlags|log.LUTC|log.Lmsgprefix)

	addr := envOrDefault("IX_SUSTAINMENT_OS_ADDR", defaultAddr)
	version := envOrDefault("IX_SUSTAINMENT_OS_VERSION", "dev")
	environment := envOrDefault("IX_SUSTAINMENT_OS_ENV", "local")

	mux := http.NewServeMux()
	mux.HandleFunc("/", rootHandler(version, environment))
	mux.HandleFunc("/healthz", healthHandler)
	mux.HandleFunc("/readyz", readinessHandler)
	mux.HandleFunc("/version", versionHandler(version, environment))

	server := &http.Server{
		Addr:         addr,
		Handler:      requestLoggingMiddleware(logger, mux),
		ReadTimeout:  defaultReadTimeout,
		WriteTimeout: defaultWriteTimeout,
		IdleTimeout:  defaultIdleTimeout,
	}

	logger.Printf("starting server on %s (env=%s version=%s)", addr, environment, version)

	serverErrCh := make(chan error, 1)
	go func() {
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrCh <- err
			return
		}
		serverErrCh <- nil
	}()

	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-stopCh:
		logger.Printf("shutdown signal received: %s", sig.String())

		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logger.Printf("graceful shutdown failed: %v", err)
			if closeErr := server.Close(); closeErr != nil {
				logger.Printf("forced close failed: %v", closeErr)
			}
			os.Exit(1)
		}

		logger.Printf("server stopped cleanly")
	case err := <-serverErrCh:
		if err != nil {
			logger.Printf("server failed: %v", err)
			os.Exit(1)
		}
	}
}

func rootHandler(version, environment string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		payload := appInfo{
			Name:        "IX Sustainment OS",
			Description: "CUI-conscious sustainment operating layer for maintenance, readiness, parts bottlenecks, technical-data access, and auditable AI-assisted decision support.",
			Version:     version,
			Environment: environment,
			Status:      "running",
			TimeUTC:     time.Now().UTC().Format(time.RFC3339),
		}

		writeJSON(w, http.StatusOK, payload)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	payload := healthResponse{
		Status:  "ok",
		Service: "ix-sustainment-os",
		TimeUTC: time.Now().UTC().Format(time.RFC3339),
	}

	writeJSON(w, http.StatusOK, payload)
}

func readinessHandler(w http.ResponseWriter, r *http.Request) {
	payload := healthResponse{
		Status:  "ready",
		Service: "ix-sustainment-os",
		TimeUTC: time.Now().UTC().Format(time.RFC3339),
	}

	writeJSON(w, http.StatusOK, payload)
}

func versionHandler(version, environment string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		payload := map[string]string{
			"service":     "ix-sustainment-os",
			"version":     version,
			"environment": environment,
			"time_utc":    time.Now().UTC().Format(time.RFC3339),
		}

		writeJSON(w, http.StatusOK, payload)
	}
}

func requestLoggingMiddleware(logger *log.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now().UTC()

		recorder := &statusRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(recorder, r)

		logger.Printf(
			"method=%s path=%s status=%d remote=%s duration_ms=%d",
			r.Method,
			r.URL.Path,
			recorder.statusCode,
			r.RemoteAddr,
			time.Since(started).Milliseconds(),
		)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(payload); err != nil {
		http.Error(w, `{"error":"failed to encode response"}`, http.StatusInternalServerError)
	}
}

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

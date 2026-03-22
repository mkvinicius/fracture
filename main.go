package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fracture/fracture/api"
	"github.com/fracture/fracture/db"
	"github.com/fracture/fracture/security"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

//go:embed dashboard/dist
var dashboardFS embed.FS

func main() {
	// ── Logger ──────────────────────────────────────────────────────────────
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	log.Info().Msg("FRACTURE starting...")

	// ── Database ────────────────────────────────────────────────────────────
	database, err := db.Open()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to open database")
	}
	defer database.Close()
	log.Info().Msg("database ready")

	// ── Security ────────────────────────────────────────────────────────────
	signerSecret := []byte(getOrGenerateSecret(database))
	signer, err := security.NewSigner(signerSecret)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create signer")
	}
	sanitizer := security.NewSanitizer(nil) // nil = regex-only mode initially
	auditLogger := security.NewAuditLogger(database.DB, signer)

	// ── Router ──────────────────────────────────────────────────────────────
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(120 * time.Second))

	// API routes
	apiHandler := api.NewHandler(database, signer, sanitizer, auditLogger)
	r.Mount("/api", apiHandler.Routes())

	// Serve dashboard (embedded React build)
	dashSub, err := fs.Sub(dashboardFS, "dashboard/dist")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to sub dashboard FS")
	}
	fileServer := http.FileServer(http.FS(dashSub))
	r.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// SPA fallback: serve index.html for unknown routes
		_, statErr := fs.Stat(dashSub, req.URL.Path[1:])
		if os.IsNotExist(statErr) {
			req.URL.Path = "/"
		}
		fileServer.ServeHTTP(w, req)
	}))

	// ── Find available port ──────────────────────────────────────────────────
	port := findAvailablePort(3000)
	addr := fmt.Sprintf("127.0.0.1:%d", port)

	// ── HTTP Server ──────────────────────────────────────────────────────────
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// ── Graceful shutdown ────────────────────────────────────────────────────
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		url := fmt.Sprintf("http://localhost:%d", port)
		log.Info().Str("url", url).Msg("FRACTURE dashboard ready")
		openBrowser(url)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	<-ctx.Done()
	log.Info().Msg("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("shutdown error")
	}
	log.Info().Msg("FRACTURE stopped")
}

// findAvailablePort finds an available port starting from preferred.
func findAvailablePort(preferred int) int {
	for port := preferred; port < preferred+100; port++ {
		ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err == nil {
			ln.Close()
			return port
		}
	}
	return preferred
}

// getOrGenerateSecret retrieves or creates the HMAC signing secret.
func getOrGenerateSecret(database *db.DB) string {
	secret, err := database.GetConfig("hmac_secret")
	if err != nil || secret == "" {
		// Generate new secret
		b := make([]byte, 32)
		if _, err := fmt.Sscanf(fmt.Sprintf("%d", time.Now().UnixNano()), "%d", new(int64)); err == nil {
			// Use time-based seed as fallback
		}
		_ = b
		newSecret := fmt.Sprintf("fracture-%d", time.Now().UnixNano())
		_ = database.SetConfig("hmac_secret", newSecret)
		return newSecret
	}
	return secret
}

// openBrowser opens the default browser to the given URL.
func openBrowser(url string) {
	// Platform-specific open — handled by build tags in separate files
	// For now, just log the URL
	log.Info().Str("url", url).Msg("open this URL in your browser")
}

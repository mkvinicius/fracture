package main

import (
	"context"
	"crypto/rand"
	"embed"
	"encoding/hex"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/fracture/fracture/api"
	"github.com/fracture/fracture/db"
	"github.com/fracture/fracture/security"
	"github.com/fracture/fracture/telemetry"
	"github.com/fracture/fracture/updater"
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
	log.Info().Str("version", updater.CurrentVersion).Msg("FRACTURE starting...")

	// ── Database ────────────────────────────────────────────────────────────
	database, err := db.Open()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to open database")
	}
	defer database.Close()
	log.Info().Msg("database ready")

	// ── Telemetria (opt-in anônima) ──────────────────────────────────────────
	telemetryURL, _ := database.GetConfig("telemetry_url")
	dataDir, _ := db.DataDir()
	// Use the canonical version from updater — single source of truth
	tel := telemetry.New(dataDir, telemetryURL, updater.CurrentVersion)
	if tel.IsEnabled() {
		tel.SendPing()
		log.Info().Msg("telemetry ping sent (opt-in enabled)")
	}

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

	// API routes — canonical prefix is /api/v1
	apiHandler := api.NewHandler(database, signer, sanitizer, auditLogger, tel)
	r.Mount("/api/v1", apiHandler.Routes())

	// Backward compat: redirect /api/<path> → /api/v1/<path> (308 preserves method)
	r.HandleFunc("/api/*", func(w http.ResponseWriter, req *http.Request) {
		suffix := strings.TrimPrefix(req.URL.Path, "/api")
		target := "/api/v1" + suffix
		if req.URL.RawQuery != "" {
			target += "?" + req.URL.RawQuery
		}
		http.Redirect(w, req, target, http.StatusPermanentRedirect) // 308
	})

	// Serve dashboard (embedded React build)
	dashSub, err := fs.Sub(dashboardFS, "dashboard/dist")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to sub dashboard FS")
	}
	fileServer := http.FileServer(http.FS(dashSub))
	r.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// SPA fallback: serve index.html for unknown routes
		path := req.URL.Path[1:]
		_, statErr := fs.Stat(dashSub, path)
		if os.IsNotExist(statErr) {
			req.URL.Path = "/"
			path = ""
		}
		// Prevent caching of index.html so browser always gets fresh asset references
		if path == "" || path == "index.html" {
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
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
		// Small delay so the server is listening before the browser opens
		time.Sleep(300 * time.Millisecond)
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
// Uses crypto/rand for cryptographically strong 256-bit secret generation.
func getOrGenerateSecret(database *db.DB) string {
	secret, err := database.GetConfig("hmac_secret")
	if err != nil || secret == "" {
		b := make([]byte, 32) // 256-bit secret
		if _, err := rand.Read(b); err != nil {
			// Extremely unlikely — only if OS entropy pool is unavailable
			log.Fatal().Err(err).Msg("failed to generate HMAC secret: crypto/rand unavailable")
		}
		newSecret := hex.EncodeToString(b)
		if err := database.SetConfig("hmac_secret", newSecret); err != nil {
			log.Warn().Err(err).Msg("failed to persist HMAC secret")
		}
		log.Info().Msg("new HMAC secret generated (crypto/rand, 256-bit)")
		return newSecret
	}
	return secret
}

// openBrowser opens the default browser to the given URL.
// Supports Linux (xdg-open), macOS (open) and Windows (rundll32).
// Non-fatal: headless/SSH environments without a browser will just log the URL.
func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default: // linux and others
		cmd = exec.Command("xdg-open", url)
	}
	if err := cmd.Start(); err != nil {
		// Non-fatal: headless environments (CI, SSH) won't have a browser
		log.Info().Str("url", url).Msg("FRACTURE ready — open this URL in your browser")
		return
	}
	log.Info().Str("url", url).Msg("FRACTURE dashboard opened in browser")
}

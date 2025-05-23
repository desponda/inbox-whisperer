package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"runtime"

	"github.com/desponda/inbox-whisperer/internal/api"
	"github.com/desponda/inbox-whisperer/internal/config"
	"github.com/desponda/inbox-whisperer/internal/data"
	"github.com/desponda/inbox-whisperer/internal/service"
	"github.com/desponda/inbox-whisperer/internal/session"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// zerologMiddleware logs each HTTP request using zerolog
func zerologMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Info().Str("method", r.Method).Str("path", r.URL.Path).Dur("duration", time.Since(start)).Msg("http request")
	})
}

func main() {
	cfg := mustLoadConfig()
	setupLogger(cfg)

	buildSHA := os.Getenv("GIT_COMMIT")
	if buildSHA == "" {
		buildSHA = "unknown"
	}
	versionMsg := "*** BACKEND VERSION INFO *** sha=" + buildSHA + " go=" + runtime.Version() + " time=" + time.Now().Format(time.RFC3339)
	log.Info().Str("build_sha", buildSHA).
		Str("go_version", runtime.Version()).
		Time("startup_time", time.Now()).
		Msg(versionMsg)
	fmt.Println(versionMsg)
	fmt.Fprintln(os.Stderr, versionMsg)

	log.Info().Msg("Starting Inbox Whisperer server")

	db := mustConnectDB(cfg)
	defer db.Close()
	log.Info().Msg("Database connection established")

	r := setupRouter(db, cfg)
	srv := setupServer(cfg, r)

	setupGracefulShutdown(srv)

	log.Info().Msgf("Server is ready to handle requests at :%s", cfg.Server.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg("Could not listen")
	}
}

func mustLoadConfig() *config.AppConfig {
	configPath := os.Getenv("CONFIG_FILE")
	if configPath == "" {
		configPath = "config.json"
	}
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
		os.Exit(1)
	}
	return cfg
}

func setupLogger(cfg *config.AppConfig) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	if cfg != nil && cfg.Server.LogLevel != "" {
		if level, err := zerolog.ParseLevel(cfg.Server.LogLevel); err == nil {
			zerolog.SetGlobalLevel(level)
		} else {
			log.Warn().Str("level", cfg.Server.LogLevel).Msg("Invalid log level, using default")
		}
	}
}

func mustConnectDB(cfg *config.AppConfig) *data.DB {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	db, err := data.New(ctx, cfg.Server.DBUrl)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	return db
}

func setupRouter(db *data.DB, cfg *config.AppConfig) http.Handler {
	r := chi.NewRouter()
	r.Use(zerologMiddleware)
	// Session middleware
	r.Use(session.Middleware)

	// Register OAuth2 endpoints
	api.RegisterAuthRoutes(r, cfg, db)
	// Only register Gmail API endpoints if db is not nil (prevents nil pointer dereference in tests)
	if db != nil {
		// Apply Auth and Token middleware to email API
		r.With(api.AuthMiddleware, api.TokenMiddleware(db)).Route("/api/email", func(r chi.Router) {
			r.Get("/messages", api.NewEmailHandler(service.NewMultiProviderEmailService(service.NewEmailProviderFactory()), db).FetchMessagesHandler)
			r.Get("/messages/{id}", api.NewEmailHandler(service.NewMultiProviderEmailService(service.NewEmailProviderFactory()), db).GetMessageContentHandler)
		})
	}

	h := api.NewUserHandler(service.NewUserService(db))
	r.Route("/users", func(r chi.Router) {
		r.Get("/", h.ListUsers)
		r.Post("/", h.CreateUser)
		// Only allow users to access/modify their own info (now via AuthMiddleware)
		r.With(api.AuthMiddleware).Get("/{id}", h.GetUser)
		r.With(api.AuthMiddleware).Put("/{id}", h.UpdateUser)
		r.With(api.AuthMiddleware).Delete("/{id}", h.DeleteUser)
	})

	// Register /api/users/me endpoint for current user info
	r.Route("/api", func(r chi.Router) {
		r.With(api.AuthMiddleware).Get("/users/me", h.GetMe)
	})

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	return r
}

func setupServer(cfg *config.AppConfig, handler http.Handler) *http.Server {
	return &http.Server{
		Handler:      handler,
		Addr:         ":" + cfg.Server.Port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
}

func setupGracefulShutdown(srv *http.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		log.Info().Msg("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Error().Err(err).Msg("Server forced to shutdown")
		}
	}()
}

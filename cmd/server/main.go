package main

import (
	"context"
	"io"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/desponda/inbox-whisperer/internal/api"
	"github.com/desponda/inbox-whisperer/internal/auth/middleware"
	"github.com/desponda/inbox-whisperer/internal/auth/service/session"
	"github.com/desponda/inbox-whisperer/internal/config"
	"github.com/desponda/inbox-whisperer/internal/data"
	"github.com/desponda/inbox-whisperer/internal/data/session/postgres"
	"github.com/desponda/inbox-whisperer/internal/service"
	"github.com/desponda/inbox-whisperer/internal/service/email"
	"github.com/desponda/inbox-whisperer/internal/service/provider"
	"github.com/desponda/inbox-whisperer/internal/service/providers/gmail"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"golang.org/x/oauth2"
)

var logger zerolog.Logger

// zerologMiddleware logs each HTTP request using zerolog
func zerologMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		logger.Info().Str("method", r.Method).Str("path", r.URL.Path).Dur("duration", time.Since(start)).Msg("http request")
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
	logger.Info().Str("build_sha", buildSHA).
		Str("go_version", runtime.Version()).
		Time("startup_time", time.Now()).
		Msg(versionMsg)

	logger.Info().Msg("Starting Inbox Whisperer server")

	db := mustConnectDB(cfg)
	defer db.Close()
	logger.Info().Msg("Database connection established")

	r := setupRouter(db, cfg)
	srv := setupServer(cfg, r)

	setupGracefulShutdown(srv)

	logger.Info().Msgf("Server is ready to handle requests at :%s", cfg.Server.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal().Err(err).Msg("Server failed to start")
	}
}

func mustLoadConfig() *config.AppConfig {
	configPath := os.Getenv("CONFIG_FILE")
	if configPath == "" {
		configPath = "config.json"
	}
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to load config")
		os.Exit(1)
	}
	return cfg
}

func setupLogger(cfg *config.AppConfig) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	var output io.Writer = os.Stdout
	if os.Getenv("LOG_FORMAT") == "console" {
		output = zerolog.ConsoleWriter{Out: os.Stdout}
	}
	logger = zerolog.New(output).With().Timestamp().Caller().Logger()
	if cfg != nil && cfg.Server.LogLevel != "" {
		if level, err := zerolog.ParseLevel(cfg.Server.LogLevel); err == nil {
			zerolog.SetGlobalLevel(level)
		} else {
			logger.Warn().Str("level", cfg.Server.LogLevel).Msg("Invalid log level, using default")
		}
	}
}

func mustConnectDB(cfg *config.AppConfig) *data.DB {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	db, err := data.New(ctx, cfg.Server.DBUrl)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to database")
	}
	return db
}

func setupRouter(db *data.DB, cfg *config.AppConfig) http.Handler {
	r := chi.NewRouter()
	r.Use(zerologMiddleware)

	// Define public paths
	publicPaths := []string{
		"/healthz",
		"/",
		"/api/auth/*",
		"/logout",
		"/api/health/*",
		"/api/docs/*",
	}

	// Create OAuth config
	oauthConfig := &oauth2.Config{
		ClientID:     cfg.Google.ClientID,
		ClientSecret: cfg.Google.ClientSecret,
		RedirectURL:  cfg.Google.RedirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/gmail.readonly",
		},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
		},
	}

	// Create session store and manager
	sessionStore, err := postgres.NewStore(db.Pool, "sessions", 24*time.Hour)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create session store")
		os.Exit(1)
	}
	sessionManager := session.NewManagerWithSecure(sessionStore, cfg.Server.SessionCookieSecure)
	defer sessionManager.Close()

	// Create middleware with public paths and OAuth config
	sessionMiddleware := middleware.NewSessionMiddleware(sessionManager, publicPaths...)
	oauthMiddleware := middleware.NewOAuthMiddleware(sessionManager, db, oauthConfig)

	// Initialize provider factory and register Gmail provider
	providerFactory := provider.NewProviderFactory()
	providerFactory.RegisterProvider(provider.Gmail, func(cfg provider.Config) (provider.Provider, error) {
		repo := data.NewEmailMessageRepositoryFromPool(db.Pool)
		// Pass oauthConfig to Gmail message service for secure API access
		gmailService := gmail.NewMessageService(repo, oauthConfig)
		return gmail.NewProvider(gmailService), nil
	})

	// Initialize handlers
	emailService := email.NewMultiProviderService(providerFactory)
	userService := service.NewUserService(db)
	emailHandler := api.NewEmailHandler(emailService, sessionManager)
	userHandler := api.NewUserHandler(userService, sessionManager)
	authHandler := api.NewAuthHandler(api.AuthHandlerDeps{
		Config:           cfg,
		UserTokens:       db,
		UserRepo:         db,
		UserIdentityRepo: db,
		SessionManager:   sessionManager,
	})

	// Public routes
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, cfg.Server.FrontendURL, http.StatusFound)
	})
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r.Get("/api/auth/login", authHandler.HandleLogin)
	r.Get("/api/auth/callback", authHandler.HandleCallback)
	r.Get("/api/logout", func(w http.ResponseWriter, r *http.Request) {
		if err := sessionManager.Destroy(w, r); err != nil {
			logger.Error().Err(err).Msg("Failed to destroy session")
		}
		http.Redirect(w, r, cfg.Server.FrontendURL, http.StatusFound)
	})

	// Protected routes
	r.Route("/api", func(r chi.Router) {
		// Apply session and OAuth middleware to all API routes
		r.Use(sessionMiddleware.Handler)
		r.Use(oauthMiddleware.Handler)

		// Email routes
		r.Route("/email", func(r chi.Router) {
			r.Get("/", emailHandler.ListEmails)
			r.Get("/{id}", emailHandler.GetEmail)
		})

		// User routes
		r.Route("/users", func(r chi.Router) {
			r.Get("/", userHandler.ListUsers)
			r.Post("/", userHandler.CreateUser)
			r.Get("/me", userHandler.GetMe)
			r.Get("/{id}", userHandler.GetUser)
			r.Put("/{id}", userHandler.UpdateUser)
			r.Delete("/{id}", userHandler.DeleteUser)
		})
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
		logger.Info().Msg("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			logger.Error().Err(err).Msg("Server forced to shutdown")
		}
	}()
}

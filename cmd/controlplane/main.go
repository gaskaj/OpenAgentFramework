package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/gaskaj/OpenAgentFramework/web/auth"
	webconfig "github.com/gaskaj/OpenAgentFramework/web/config"
	"github.com/gaskaj/OpenAgentFramework/web/handler"
	"github.com/gaskaj/OpenAgentFramework/web/migrate"
	"github.com/gaskaj/OpenAgentFramework/web/router"
	"github.com/gaskaj/OpenAgentFramework/web/server"
	"github.com/gaskaj/OpenAgentFramework/web/store"
	"github.com/gaskaj/OpenAgentFramework/web/ws"
)

func main() {
	configPath := flag.String("config", "configs/controlplane.yaml", "path to config file")
	flag.Parse()

	// Setup logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Load config
	cfg, err := webconfig.Load(*configPath)
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// Set log level
	switch cfg.Logging.Level {
	case "debug":
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case "warn":
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	case "error":
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Initialize database
	logger.Info("connecting to database", "host", cfg.Database.Host, "name", cfg.Database.Name)
	stores, err := store.NewPostgresStore(ctx, cfg.Database, logger, nil)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer stores.Close()

	// Run migrations
	logger.Info("running database migrations")
	if err := migrate.Run(ctx, stores.Pool, logger); err != nil {
		logger.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	// Initialize auth
	jwtMgr := auth.NewJWTManager(cfg.Auth.JWTSecret, cfg.Auth.JWTExpiry, cfg.Auth.RefreshExpiry)

	// Initialize WebSocket hub
	hub := ws.NewHub(logger)

	// Initialize auth handler with OAuth providers
	authHandler := handler.NewAuthHandler(stores.Users, stores.Orgs, jwtMgr, cfg.Auth.BcryptCost, logger)
	if cfg.Auth.Google.Enabled {
		authHandler.RegisterProvider("google", auth.NewGoogleProvider(cfg.Auth.Google))
	}
	if cfg.Auth.Azure.Enabled {
		authHandler.RegisterProvider("azure", auth.NewAzureProvider(cfg.Auth.Azure))
	}

	// Create router
	r := router.New(stores, jwtMgr, hub, authHandler, cfg.Versioning, logger, cfg.CORS.AllowedOrigins)

	// Create and start server
	srv := server.New(
		cfg.Server.Addr(),
		r,
		cfg.Server.ReadTimeout,
		cfg.Server.WriteTimeout,
		logger,
	)

	// Start server in background
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start()
	}()

	logger.Info("control plane server started", "addr", cfg.Server.Addr())

	// Wait for shutdown signal or server error
	select {
	case err := <-errCh:
		if err != nil {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	case <-ctx.Done():
		logger.Info("shutdown signal received")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
		defer shutdownCancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Error("server shutdown error", "error", err)
		}
	}

	logger.Info("control plane server stopped")
}

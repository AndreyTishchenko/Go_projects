package main

import (
	"context"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	articlepostgres "github.com/AndreyTishchenko/Go_projects/personal_blog/internal/articles/adapters/postgres"
	"github.com/AndreyTishchenko/Go_projects/personal_blog/internal/articles/app"
	articlehttp "github.com/AndreyTishchenko/Go_projects/personal_blog/internal/articles/http"
	"github.com/AndreyTishchenko/Go_projects/personal_blog/internal/config"
	platformlog "github.com/AndreyTishchenko/Go_projects/personal_blog/internal/platform/log"
	platformpostgres "github.com/AndreyTishchenko/Go_projects/personal_blog/internal/platform/postgres"
	"github.com/AndreyTishchenko/Go_projects/personal_blog/internal/platform/session"
)

func main() {
	logger := platformlog.New()
	if err := run(logger); err != nil {
		logger.Error("application stopped with an error", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	connectCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	db, err := platformpostgres.Open(connectCtx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	repository := articlepostgres.NewRepository(db)
	articles := app.NewService(repository)
	sessions := session.NewManager(cfg.AdminLogin, cfg.AdminPasswordHash, cfg.SessionTTL)
	templates := template.Must(template.ParseGlob("templates/*.html"))
	handler := articlehttp.NewHandler(articles, templates, sessions, logger)

	server := http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler.Routes(),
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("application starting", "address", server.Addr)
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("serve HTTP: %w", err)
		}
	case <-ctx.Done():
		logger.Info("shutdown requested")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		_ = server.Close()
		return fmt.Errorf("graceful shutdown: %w", err)
	}
	logger.Info("application stopped")
	return nil
}

package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
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
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load configuration: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := platformpostgres.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	logger := platformlog.New()
	repository := articlepostgres.NewRepository(db)
	articles := app.NewService(repository)
	sessions := session.NewManager(cfg.AdminLogin, cfg.AdminPassword)
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

	logger.Info("application running", "address", server.Addr)
	log.Fatal(server.ListenAndServe())
}

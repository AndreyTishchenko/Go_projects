package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/AndreyTishchenko/Go_projects/personal_blog/repository"
	"github.com/AndreyTishchenko/Go_projects/personal_blog/server"
)

func main() {
	portNumber := ":" + os.Getenv("PORT")
	if portNumber == ":" {
		portNumber = ":8080"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("create database pool: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		log.Fatalf("connect to database: %v", err)
	}

	log.Println("connected to PostgreSQL")

	tmpl := template.Must(template.ParseGlob("templates/*.html"))

	articlesRepository := repository.NewArticlesPostgresRepository(db)

	s := server.NewServerConfig(
		articlesRepository,
		tmpl,
	)

	httpServer := http.Server{
		Addr:              portNumber,
		Handler:           s.Routes(),
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("application running on %s", portNumber)
	log.Fatal(httpServer.ListenAndServe())
}

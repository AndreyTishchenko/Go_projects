package main

import (
	"crypto/rand"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/AndreyTishchenko/Go_projects/personal_blog/repository"
	"github.com/AndreyTishchenko/Go_projects/personal_blog/server"
)

const portNumber = ":8080"

func main() {
	tmpl := template.Must(template.ParseGlob("templates/*.html"))

	s, err := server.NewServerConfig(&repository.ArticlesMemoryRepository{
		DbPath: "db/articles/",
	}, tmpl, rand.Reader)
	if err != nil {
		log.Fatal("failed to create server config: ", err)
	}

	http_server := http.Server{
		Addr:              "localhost:8080",
		Handler:           s.Routes(),
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	fmt.Println("Application running on %r", portNumber)
	log.Fatal(http_server.ListenAndServe())
}

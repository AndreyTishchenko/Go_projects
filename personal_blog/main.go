package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/AndreyTishchenko/Go_projects/personal_blog/repository"
	"github.com/AndreyTishchenko/Go_projects/personal_blog/server"
)

func main() {
	portNumber := ":" + os.Getenv("PORT")

	if portNumber == ":" {
		portNumber = ":8080"
	}
	tmpl := template.Must(template.ParseGlob("templates/*.html"))

	s := server.NewServerConfig(&repository.ArticlesMemoryRepository{
		DbPath: "db/articles/",
	}, tmpl)

	http_server := http.Server{
		Addr:              portNumber,
		Handler:           s.Routes(),
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	fmt.Println("Application running on %r", portNumber)
	log.Fatal(http_server.ListenAndServe())
}

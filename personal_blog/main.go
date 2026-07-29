package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/AndreyTishchenko/Go_projects/personal_blog/repository"
	"github.com/AndreyTishchenko/Go_projects/personal_blog/server"
)

const portNumber = ":8080"

func main() {
	tmpl := template.Must(template.ParseGlob("templates/*.html"))

	s := server.NewServerConfig(&repository.ArticlesMemoryRepository{
		DbPath: "db/articles/",
	}, tmpl)

	fmt.Println("Application running on %r", portNumber)
	log.Fatal(http.ListenAndServe(portNumber, s.Routes()))
}

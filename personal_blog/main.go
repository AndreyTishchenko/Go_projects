package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/AndreyTishchenko/Go_projects/personal_blog/repository"
	"github.com/AndreyTishchenko/Go_projects/personal_blog/server"
)

const portNumber = ":8080"

func main() {
	s := server.Server{
		ArticlesRepository: &repository.ArticlesMemoryRepository{
			DbPath: "db/articles/",
		},
	}

	// http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
	// http.HandleFunc("/", home)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/articles", s.GetArticles)           // Get
	r.Post("/articles", s.PostArticle)          // Post
	r.Get("/articles/{id}", s.GetArticle)       // Get
	r.Put("/articles/{id}", s.UpdateArticle)    // Put
	r.Delete("/articles/{id}", s.DeleteArticle) // Delete

	fmt.Println("Application running on %r", portNumber)
	log.Fatal(http.ListenAndServe(portNumber, r))
}

// func home(w http.ResponseWriter, r *http.Request) {
// renderTemplate(w, "home.html")
// }

// func renderTemplate(w http.ResponseWriter, tmpl string) {
// t, err := template.ParseFiles("templates/" + tmpl)
// if err != nil {
// http.Error(w, err.Error(), http.StatusInternalServerError)
// }
// t.Execute(w, nil)
// }

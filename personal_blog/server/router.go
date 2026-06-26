package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (s Server) AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("auth")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		isAdmin := s.AuthCheck(cookie.Value)

		if !isAdmin {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s Server) Routes() http.Handler {
	r := chi.NewRouter()

	// static files
	fs := http.FileServer(http.Dir("./static"))

	r.Handle("/static/*", http.StripPrefix("/static/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		fs.ServeHTTP(w, r)
	})))

	// API
	r.Route("/api", func(api chi.Router) {
		api.Get("/articles", s.GetArticles)
		api.Get("/articles/{id}", s.GetArticle)
	})

	r.Get("/", s.HomePage)
	r.Get("/article/{id}", s.ArticlePage)
	r.Get("/login", s.LoginPage)
	r.Post("/auth", s.AuthHandler)

	r.Route("/admin", func(api chi.Router) {
		api.Use(s.AdminOnly)

		api.Get("/", s.AdminPage)
		api.Post("/change/{id}", s.PostArticle)
		api.Post("/delete/{id}", s.DeleteArticle)
		api.Get("/new", s.GetArticle)
	})

	return r
}

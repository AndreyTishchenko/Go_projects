package http

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

func (s *Handler) AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("auth")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		isAdmin := s.sessions.IsAdmin(cookie.Value)

		if !isAdmin {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Use(s.RequestLogger)

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK\n"))
	})

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
		api.Post("/articles", s.PostArticle)
	})

	r.Get("/", s.HomePage)
	r.Get("/article/{id}", s.ArticlePage)
	r.Get("/login", s.LoginPage)
	r.Post("/auth", s.AuthHandler)
	r.Post("/logout", s.LogoutHandler)

	r.Route("/admin", func(api chi.Router) {
		api.Use(s.AdminOnly)

		api.Get("/", s.AdminPage)
		api.Get("/change/{id}", s.ChangeArticle)
		api.Post("/change/{id}", s.UpdateArticle)
		api.Post("/delete/{id}", s.DeleteArticle)
		api.Get("/new", s.AddArticle)
		api.Post("/new", s.PostArticle)
	})

	return r
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (w *statusRecorder) WriteHeader(status int) {
	if w.status != 0 {
		return
	}
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusRecorder) Write(p []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(p)
}

func (w *statusRecorder) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func (s *Handler) RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &statusRecorder{ResponseWriter: w}

		next.ServeHTTP(recorder, r)

		status := recorder.status
		if status == 0 {
			status = http.StatusOK
		}
		s.logger.Info("request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"status", status,
			"duration", time.Since(start),
		)
	})
}

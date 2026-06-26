package server

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/AndreyTishchenko/Go_projects/personal_blog/repository"
	"github.com/go-chi/chi/v5"
)

func (s Server) GetArticles(w http.ResponseWriter, r *http.Request) {
	articles, err := s.ArticlesRepository.GetArticles()

	if err != nil {
		log.Println("Failed to read payload:", err.Error())
		http.Error(w, "Failed to read payload", http.StatusInternalServerError)
		return
	}

	jsonArticles, err := json.Marshal(articles)

	if err != nil {
		log.Println("Failed to get articles:", err.Error())
		http.Error(w, "Failed to get articles", http.StatusInternalServerError)
		return
	}

	w.Write(jsonArticles)
}

func (s Server) GetArticle(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)

	if err != nil {
		http.Error(w, "invalid article id format", http.StatusBadRequest)
		return
	}

	article, err := s.ArticlesRepository.GetArticle(id)

	if err != nil {
		if errors.Is(err, repository.ErrArticleNotFound) {
			println("Article is not found:", err)
			http.Error(w, "Article is not found", http.StatusNotFound)
			return
		}
		println("Cannot get article:", err)
		http.Error(w, "Cannot get article", http.StatusInternalServerError)
		return
	}

	jsonArticle, err := json.Marshal(article)

	if err != nil {
		println("Cannot get article:", err)
		http.Error(w, "Cannot get article", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonArticle)
}

func (s Server) PostArticle(w http.ResponseWriter, r *http.Request) {
	var article ArticlePayload

	isBrowser := r.Header.Get("Content-Type") != "application/json"

	isAdmin := false

	if isBrowser {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		article.Body = r.FormValue("body")
		article.Title = r.FormValue("title")

		http.SetCookie(w, &http.Cookie{
			Name:     "Title",
			Value:    url.QueryEscape(article.Title),
			Path:     "/",
			MaxAge:   60, // seconds
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})

		http.SetCookie(w, &http.Cookie{
			Name:     "Body",
			Value:    url.QueryEscape(article.Body),
			Path:     "/",
			MaxAge:   60, // seconds
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})

		if article.Body == "" && article.Title == "" {
			SetFlash(w, ErrBothFieldsEmpty.Error())
			http.Redirect(w, r, "/admin/new", http.StatusSeeOther)
			return
		} else if article.Body == "" {
			SetFlash(w, ErrEmptyBodyField.Error())
			http.Redirect(w, r, "/admin/new", http.StatusSeeOther)
			return
		} else if article.Title == "" {
			SetFlash(w, ErrEmptyTitleField.Error())
			http.Redirect(w, r, "/admin/new", http.StatusSeeOther)
			return
		}

		cookie, err := r.Cookie("auth")

		if err != nil {
			if err != http.ErrNoCookie {
				log.Println("cookie parse error:", err)
			}
			isAdmin = false
		} else {
			isAdmin = s.AuthCheck(cookie.Value)
		}

		if isAdmin != true {
			http.Redirect(w, r, "/login", http.StatusForbidden)
			return
		}
	} else {
		if r.Method != http.MethodPost {
			http.Error(w, "invalid method", http.StatusMethodNotAllowed)
			return
		}

		err := json.NewDecoder(r.Body).Decode(&article)
		if err != nil {
			http.Error(w, "invalid format", http.StatusBadRequest)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "Title",
			Value:    url.QueryEscape(article.Title),
			Path:     "/",
			MaxAge:   60, // seconds
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})

		http.SetCookie(w, &http.Cookie{
			Name:     "Body",
			Value:    url.QueryEscape(article.Body),
			Path:     "/",
			MaxAge:   60, // seconds
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})

		if article.Title == "" || article.Body == "" {
			http.Error(w, "title and body are required", http.StatusBadRequest)
			return
		}

		cookie, err := r.Cookie("auth")

		if err != nil {
			if err != http.ErrNoCookie {
				log.Println("cookie parse error:", err)
				http.Error(w, "Cannot read cookies", http.StatusInternalServerError)
				return
			}
			isAdmin = false
		} else {
			isAdmin = s.AuthCheck(cookie.Value)
		}

		if isAdmin != true {
			http.Error(w, "Must be logged in", http.StatusForbidden)
			return
		}
	}

	id, err := s.ArticlesRepository.AddArticle(article.Title, article.Body)

	if isBrowser {
		if err != nil {
			println("failed to create article", err.Error())
			http.Redirect(w, r, "/admin/new", http.StatusInternalServerError)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:     "Title",
			Value:    "",
			Path:     "/",
			MaxAge:   -1, // seconds
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})

		http.SetCookie(w, &http.Cookie{
			Name:     "Body",
			Value:    "",
			Path:     "/",
			MaxAge:   -1, // seconds
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	} else {
		if err != nil {
			println("failed to create article", err.Error())
			http.Error(w, "failed to create article", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "Title",
			Value:    "",
			Path:     "/",
			MaxAge:   -1, // seconds
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})

		http.SetCookie(w, &http.Cookie{
			Name:     "Body",
			Value:    "",
			Path:     "/",
			MaxAge:   -1, // seconds
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]int{
			"id": id,
		})
	}
}

func (s Server) UpdateArticle(w http.ResponseWriter, r *http.Request) {
	var article ArticlePayload

	err := json.NewDecoder(r.Body).Decode(&article)

	if err != nil {
		println("invalid request body ", err.Error())
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if article.Title == "" || article.Body == "" {
		http.Error(w, "title and body are required", http.StatusBadRequest)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)

	if err != nil {
		http.Error(w, "invalid article id format", http.StatusBadRequest)
		return
	}

	err = s.ArticlesRepository.UpdateArticle(id, article.Title, article.Body)

	if err != nil {
		if errors.Is(err, repository.ErrArticleNotFound) {
			println("Article is not found:", err)
			http.Error(w, "Article is not found", http.StatusNotFound)
			return
		}
		println("Cannot get article:", err)
		http.Error(w, "Cannot change article", http.StatusInternalServerError)
		return
	}
}

func (s Server) DeleteArticle(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)

	if err != nil {
		http.Error(w, "invalid article id format", http.StatusBadRequest)
		return
	}

	isAdmin := false

	cookie, err := r.Cookie("auth")

	if err != nil {
		if err != http.ErrNoCookie {
			log.Println("cookie parse error:", err)
		}
		isAdmin = false
	} else {
		isAdmin = s.AuthCheck(cookie.Value)
	}

	if isAdmin != true {
		http.Redirect(w, r, "/login", http.StatusForbidden)
		return
	}

	err = s.ArticlesRepository.DeleteArticle(id)

	if err != nil {
		if errors.Is(err, repository.ErrArticleNotFound) {
			println("Article is not found:", err)
			http.Error(w, "Article is not found", http.StatusNotFound)
			return
		}
		println("Cannot get article:", err)
		http.Error(w, "Cannot get article", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

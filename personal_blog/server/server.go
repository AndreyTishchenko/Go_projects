package server

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/AndreyTishchenko/Go_projects/personal_blog/repository"
	"github.com/go-chi/chi/v5"
)

type Server struct {
	ArticlesRepository repository.ArticlesRepository
}

type ArticlePayload struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

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

	id, err := s.ArticlesRepository.AddArticle(article.Title, article.Body)
	if err != nil {
		println("failed to create article", err.Error())
		http.Error(w, "failed to create article", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(map[string]int{
		"id": id,
	})
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

	w.WriteHeader(http.StatusNoContent)
}

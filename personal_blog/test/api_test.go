package test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/AndreyTishchenko/Go_projects/personal_blog/repository"
)

func TestAPIGetArticleResponses(t *testing.T) {
	article := repository.Article{
		ID:        2,
		CreatedAt: time.Date(2026, 7, 2, 18, 25, 12, 0, time.UTC),
		Title:     "API article",
		Text:      "API body",
	}

	_, handler := newTestApplication(t, newFakeArticlesRepository(article))

	rr := performRequest(handler, http.MethodGet, "/api/articles/2", "")
	if rr.Code != http.StatusOK {
		t.Fatalf("GET /api/articles/2 status = %d, want %d", rr.Code, http.StatusOK)
	}

	var got repository.Article
	if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
		t.Fatalf("decode GET /api/articles/2 response: %v", err)
	}

	if got != article {
		t.Fatalf("GET /api/articles/2 article = %v, want %v", got, article)
	}

	rr = performRequest(handler, http.MethodGet, "/api/articles/bad-id", "")
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("GET /api/articles/bad-id status = %d, want %d", rr.Code, http.StatusBadRequest)
	}

	rr = performRequest(handler, http.MethodGet, "/api/articles/404", "")
	if rr.Code != http.StatusNotFound {
		t.Fatalf("GET /api/articles/404 status = %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestJSONAuthAndCreateArticle(t *testing.T) {
	repo := newFakeArticlesRepository()
	_, handler := newTestApplication(t, repo)
	authCookie := loginAsAdmin(t, handler)

	rr := performRequest(handler, http.MethodPost, "/api/articles", `{"title":"New article","body":"New body"}`, authCookie)

	if rr.Code != http.StatusCreated {
		t.Fatalf("POST /api/articles status = %d, want %d; body = %q", rr.Code, http.StatusCreated, rr.Body.String())
	}

	var response struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("decode POST /api/articles response: %v", err)
	}

	article, err := repo.GetArticle(response.ID)
	if err != nil {
		t.Fatalf("created article not stored: %v", err)
	}
	if article.Title != "New article" || article.Text != "New body" {
		t.Fatalf("created article = %v, want posted title/body", article)
	}
}

func TestJSONCreateArticleValidation(t *testing.T) {
	_, handler := newTestApplication(t, newFakeArticlesRepository())
	authCookie := loginAsAdmin(t, handler)

	rr := performRequest(handler, http.MethodPost, "/api/articles", `{"title":"","body":"Body"}`, authCookie)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("POST /api/articles with empty title status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
	if !strings.Contains(rr.Body.String(), "title and body are required") {
		t.Fatalf("POST /api/articles with empty title body = %q, want validation message", rr.Body.String())
	}
}

package test

import (
	"bytes"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/AndreyTishchenko/Go_projects/personal_blog/internal/articles/app"
)

func TestHomePageRendersArticles(t *testing.T) {
	article := app.Article{
		ID:        7,
		CreatedAt: time.Date(2026, 7, 2, 18, 25, 12, 0, time.UTC),
		Title:     "Application testing",
		Text:      "Handler body",
	}
	_, handler := newTestApplication(t, newFakeArticlesRepository(article))

	rr := performRequest(handler, http.MethodGet, "/", "")

	if rr.Code != http.StatusOK {
		t.Fatalf("GET / status = %d, want %d", rr.Code, http.StatusOK)
	}

	body := rr.Body.String()
	for _, want := range []string{"Personal Blog", "Application testing", "July 2, 2026", `/article/7`} {
		if !strings.Contains(body, want) {
			t.Fatalf("GET / body does not contain %q; body = %q", want, body)
		}
	}
}

func TestHomePageRepositoryError(t *testing.T) {
	repo := newFakeArticlesRepository()
	repo.err = errors.New("boom")
	_, handler := newTestApplication(t, repo)

	rr := performRequest(handler, http.MethodGet, "/", "")

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("GET / with repository error status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}
	if !bytes.Contains(rr.Body.Bytes(), []byte("Internal server error")) {
		t.Fatalf("GET / with repository error body = %q, want internal server error", rr.Body.String())
	}
}

func TestArticlePageRendersArticleAndNotFound(t *testing.T) {
	article := app.Article{
		ID:        3,
		CreatedAt: time.Date(2026, 7, 1, 10, 0, 0, 0, time.UTC),
		Title:     "Deep dive",
		Text:      "Article content",
	}
	_, handler := newTestApplication(t, newFakeArticlesRepository(article))

	rr := performRequest(handler, http.MethodGet, "/article/3", "")
	if rr.Code != http.StatusOK {
		t.Fatalf("GET /article/3 status = %d, want %d", rr.Code, http.StatusOK)
	}

	body := rr.Body.String()
	for _, want := range []string{"Deep dive", "Article content", "July 1, 2026"} {
		if !strings.Contains(body, want) {
			t.Fatalf("GET /article/3 body does not contain %q; body = %q", want, body)
		}
	}

	rr = performRequest(handler, http.MethodGet, "/article/missing", "")
	if rr.Code != http.StatusNotFound {
		t.Fatalf("GET /article/missing status = %d, want %d", rr.Code, http.StatusNotFound)
	}
	if !strings.Contains(rr.Body.String(), "Page not found") {
		t.Fatalf("GET /article/missing body = %q, want 404 page", rr.Body.String())
	}
}

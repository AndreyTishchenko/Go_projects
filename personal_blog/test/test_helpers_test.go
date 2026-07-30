package test

import (
	"context"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/AndreyTishchenko/Go_projects/personal_blog/internal/articles/app"
	articlehttp "github.com/AndreyTishchenko/Go_projects/personal_blog/internal/articles/http"
	"github.com/AndreyTishchenko/Go_projects/personal_blog/internal/platform/session"
)

type fakeArticlesRepository struct {
	articles map[int]app.Article
	nextID   int
	err      error
}

func newFakeArticlesRepository(articles ...app.Article) *fakeArticlesRepository {
	r := &fakeArticlesRepository{
		articles: map[int]app.Article{},
		nextID:   1,
	}

	for _, article := range articles {
		r.articles[article.ID] = article
		if article.ID >= r.nextID {
			r.nextID = article.ID + 1
		}
	}

	return r
}

func (r *fakeArticlesRepository) List(context.Context) ([]app.Article, error) {
	if r.err != nil {
		return nil, r.err
	}

	articles := make([]app.Article, 0, len(r.articles))
	for _, article := range r.articles {
		articles = append(articles, article)
	}

	return articles, nil
}

func (r *fakeArticlesRepository) Get(_ context.Context, id int) (app.Article, error) {
	if r.err != nil {
		return app.Article{}, r.err
	}

	article, ok := r.articles[id]
	if !ok {
		return app.Article{}, app.ErrNotFound
	}

	return article, nil
}

func (r *fakeArticlesRepository) Create(_ context.Context, title string, body string) (int, error) {
	if r.err != nil {
		return 0, r.err
	}

	id := r.nextID
	r.nextID++
	r.articles[id] = app.Article{
		ID:        id,
		CreatedAt: time.Date(2026, 7, 3, 12, 0, 0, 0, time.UTC),
		Title:     title,
		Text:      body,
	}

	return id, nil
}

func (r *fakeArticlesRepository) Delete(_ context.Context, id int) error {
	if r.err != nil {
		return r.err
	}

	if _, ok := r.articles[id]; !ok {
		return app.ErrNotFound
	}

	delete(r.articles, id)
	return nil
}

func (r *fakeArticlesRepository) Update(_ context.Context, id int, title string, body string) error {
	if r.err != nil {
		return r.err
	}

	article, ok := r.articles[id]
	if !ok {
		return app.ErrNotFound
	}

	article.Title = title
	article.Text = body
	r.articles[id] = article

	return nil
}

func newTestApplication(t *testing.T, repo app.Repository) (*articlehttp.Handler, http.Handler) {
	t.Helper()

	tmpl := template.Must(template.ParseGlob("../templates/*.html"))
	service := app.NewService(repo)
	sessions := session.NewManager("admin", "admin213")
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	s := articlehttp.NewHandler(service, tmpl, sessions, logger)

	return s, s.Routes()
}

func performRequest(handler http.Handler, method string, path string, body string, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	if body != "" && strings.HasPrefix(body, "{") {
		req.Header.Set("Content-Type", "application/json")
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	return rr
}

func loginAsAdmin(t *testing.T, handler http.Handler) *http.Cookie {
	t.Helper()

	rr := performRequest(handler, http.MethodPost, "/auth", `{"name":"admin","password":"admin213"}`)

	if rr.Code != http.StatusOK {
		t.Fatalf("POST /auth status = %d, want %d; body = %q", rr.Code, http.StatusOK, rr.Body.String())
	}

	for _, cookie := range rr.Result().Cookies() {
		if cookie.Name == "auth" {
			return cookie
		}
	}

	t.Fatal("POST /auth did not set auth cookie")
	return nil
}

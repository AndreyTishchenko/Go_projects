package test

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/AndreyTishchenko/Go_projects/personal_blog/repository"
	"github.com/AndreyTishchenko/Go_projects/personal_blog/server"
)

type fakeArticlesRepository struct {
	articles map[int]repository.Article
	nextID   int
	err      error
}

func newFakeArticlesRepository(articles ...repository.Article) *fakeArticlesRepository {
	r := &fakeArticlesRepository{
		articles: map[int]repository.Article{},
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

func (r *fakeArticlesRepository) GetArticles() ([]repository.Article, error) {
	if r.err != nil {
		return nil, r.err
	}

	articles := make([]repository.Article, 0, len(r.articles))
	for _, article := range r.articles {
		articles = append(articles, article)
	}

	return articles, nil
}

func (r *fakeArticlesRepository) GetArticle(id int) (repository.Article, error) {
	if r.err != nil {
		return repository.Article{}, r.err
	}

	article, ok := r.articles[id]
	if !ok {
		return repository.Article{}, repository.ErrArticleNotFound
	}

	return article, nil
}

func (r *fakeArticlesRepository) AddArticle(title string, body string) (int, error) {
	if r.err != nil {
		return 0, r.err
	}

	id := r.nextID
	r.nextID++
	r.articles[id] = repository.Article{
		ID:        id,
		CreatedAt: time.Date(2026, 7, 3, 12, 0, 0, 0, time.UTC),
		Title:     title,
		Text:      body,
	}

	return id, nil
}

func (r *fakeArticlesRepository) DeleteArticle(id int) error {
	if r.err != nil {
		return r.err
	}

	if _, ok := r.articles[id]; !ok {
		return repository.ErrArticleNotFound
	}

	delete(r.articles, id)
	return nil
}

func (r *fakeArticlesRepository) UpdateArticle(id int, title string, body string) error {
	if r.err != nil {
		return r.err
	}

	article, ok := r.articles[id]
	if !ok {
		return repository.ErrArticleNotFound
	}

	article.Title = title
	article.Text = body
	r.articles[id] = article

	return nil
}

func newTestApplication(t *testing.T, repo repository.ArticlesRepository) (server.Server, http.Handler) {
	t.Helper()

	tmpl := template.Must(template.ParseGlob("../templates/*.html"))
	s := server.NewServerConfig(repo, tmpl)

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

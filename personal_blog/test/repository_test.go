package test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	articlepostgres "github.com/AndreyTishchenko/Go_projects/personal_blog/internal/articles/adapters/postgres"
	"github.com/AndreyTishchenko/Go_projects/personal_blog/internal/articles/app"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/jackc/pgx/v5/pgxpool"
)

func newTestRepository(t *testing.T) *articlepostgres.Repository {
	t.Helper()

	var cfg struct {
		DatabaseURL string `env:"TEST_DATABASE_URL"`
	}
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		t.Fatalf("load test configuration: %v", err)
	}
	if cfg.DatabaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set; run the database tests with `docker compose run --rm test`")
	}

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		t.Fatalf("create test database pool: %v", err)
	}
	t.Cleanup(pool.Close)

	if err := pool.Ping(context.Background()); err != nil {
		t.Fatalf("connect to test database: %v", err)
	}

	if _, err := pool.Exec(context.Background(), "TRUNCATE articles RESTART IDENTITY"); err != nil {
		t.Fatalf("reset articles table: %v", err)
	}

	return articlepostgres.NewRepository(pool)
}

func addArticle(t *testing.T, repo *articlepostgres.Repository, title, body string) app.Article {
	t.Helper()

	id, err := repo.Create(t.Context(), title, body)
	if err != nil {
		t.Fatalf("AddArticle(%q, %q): %v", title, body, err)
	}

	article, err := repo.Get(t.Context(), id)
	if err != nil {
		t.Fatalf("GetArticle(%d): %v", id, err)
	}

	return article
}

func TestPostgresRepositoryGetArticles(t *testing.T) {
	repo := newTestRepository(t)
	first := addArticle(t, repo, "First", "First body")
	time.Sleep(time.Millisecond)
	second := addArticle(t, repo, "Second", "Second body")

	got, err := repo.List(t.Context())
	if err != nil {
		t.Fatalf("GetArticles() error = %v", err)
	}

	want := []app.Article{second, first}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("GetArticles() = %v, want newest-first %v", got, want)
	}
}

func TestPostgresRepositoryGetArticle(t *testing.T) {
	repo := newTestRepository(t)
	want := addArticle(t, repo, "Hello World", "Hello!")

	got, err := repo.Get(t.Context(), want.ID)
	if err != nil {
		t.Fatalf("GetArticle(%d) error = %v", want.ID, err)
	}
	if got != want {
		t.Fatalf("GetArticle(%d) = %v, want %v", want.ID, got, want)
	}
}

func TestPostgresRepositoryGetArticleReturnsNotFound(t *testing.T) {
	repo := newTestRepository(t)

	_, err := repo.Get(t.Context(), 404)
	if !errors.Is(err, app.ErrNotFound) {
		t.Fatalf("Get(404) error = %v, want %v", err, app.ErrNotFound)
	}
}

func TestPostgresRepositoryAddArticle(t *testing.T) {
	repo := newTestRepository(t)

	id, err := repo.Create(t.Context(), "New title", "New body")
	if err != nil {
		t.Fatalf("AddArticle() error = %v", err)
	}
	if id != 1 {
		t.Fatalf("AddArticle() id = %d, want 1 after identity reset", id)
	}

	got, err := repo.Get(t.Context(), id)
	if err != nil {
		t.Fatalf("GetArticle(%d) after AddArticle error = %v", id, err)
	}
	if got.Title != "New title" || got.Text != "New body" || got.CreatedAt.IsZero() {
		t.Fatalf("added article = %v, want title, body, and database timestamp", got)
	}
}

func TestPostgresRepositoryDeleteArticle(t *testing.T) {
	repo := newTestRepository(t)
	article := addArticle(t, repo, "Delete me", "Body")

	if err := repo.Delete(t.Context(), article.ID); err != nil {
		t.Fatalf("DeleteArticle(%d) error = %v", article.ID, err)
	}
	if _, err := repo.Get(t.Context(), article.ID); !errors.Is(err, app.ErrNotFound) {
		t.Fatalf("Get(%d) after delete error = %v, want %v", article.ID, err, app.ErrNotFound)
	}
}

func TestPostgresRepositoryDeleteArticleReturnsNotFound(t *testing.T) {
	repo := newTestRepository(t)

	err := repo.Delete(t.Context(), 404)
	if !errors.Is(err, app.ErrNotFound) {
		t.Fatalf("Delete(404) error = %v, want %v", err, app.ErrNotFound)
	}
}

func TestPostgresRepositoryUpdateArticle(t *testing.T) {
	repo := newTestRepository(t)
	article := addArticle(t, repo, "Old title", "Old body")

	if err := repo.Update(t.Context(), article.ID, "Updated title", "Updated body"); err != nil {
		t.Fatalf("UpdateArticle(%d) error = %v", article.ID, err)
	}

	got, err := repo.Get(t.Context(), article.ID)
	if err != nil {
		t.Fatalf("GetArticle(%d) after UpdateArticle error = %v", article.ID, err)
	}
	if got.ID != article.ID || got.Title != "Updated title" || got.Text != "Updated body" {
		t.Fatalf("updated article = %v, want updated title/body with same id", got)
	}
	if !got.CreatedAt.Equal(article.CreatedAt) {
		t.Fatalf("updated article CreatedAt = %v, want original %v", got.CreatedAt, article.CreatedAt)
	}
}

func TestPostgresRepositoryUpdateArticleReturnsNotFound(t *testing.T) {
	repo := newTestRepository(t)

	err := repo.Update(t.Context(), 404, "Title", "Body")
	if !errors.Is(err, app.ErrNotFound) {
		t.Fatalf("Update(404) error = %v, want %v", err, app.ErrNotFound)
	}
}

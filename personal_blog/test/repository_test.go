package test

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/AndreyTishchenko/Go_projects/personal_blog/repository"
)

func newTestRepository(t *testing.T) repository.ArticlesMemoryRepository {
	t.Helper()

	dbPath := t.TempDir()
	if err := os.WriteFile(filepath.Join(dbPath, "counter.txt"), []byte("0"), 0644); err != nil {
		t.Fatalf("write counter: %v", err)
	}

	return repository.ArticlesMemoryRepository{DbPath: dbPath}
}

func writeArticle(t *testing.T, dbPath string, article repository.Article) {
	t.Helper()

	data, err := json.Marshal(article)
	if err != nil {
		t.Fatalf("marshal article: %v", err)
	}

	path := filepath.Join(dbPath, strconv.Itoa(article.ID)+".json")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("write article: %v", err)
	}
}

func TestGetArticles(t *testing.T) {
	repo := newTestRepository(t)
	articles := []repository.Article{
		{
			ID:        2,
			CreatedAt: time.Date(2026, 7, 2, 18, 25, 12, 0, time.UTC),
			Title:     "Second",
			Text:      "Second body",
		},
		{
			ID:        1,
			CreatedAt: time.Date(2026, 7, 1, 10, 0, 0, 0, time.UTC),
			Title:     "First",
			Text:      "First body",
		},
	}

	for _, article := range articles {
		writeArticle(t, repo.DbPath, article)
	}

	got, err := repo.GetArticles()
	if err != nil {
		t.Fatalf("GetArticles() error = %v", err)
	}

	gotByID := map[int]repository.Article{}
	for _, article := range got {
		gotByID[article.ID] = article
	}

	wantByID := map[int]repository.Article{
		1: articles[1],
		2: articles[0],
	}

	if !reflect.DeepEqual(gotByID, wantByID) {
		t.Fatalf("GetArticles() = %v, want %v", gotByID, wantByID)
	}
}

func TestGetArticlesReturnsInvalidJSONError(t *testing.T) {
	repo := newTestRepository(t)
	if err := os.WriteFile(filepath.Join(repo.DbPath, "broken.json"), []byte(`{"id":1,"title":`), 0644); err != nil {
		t.Fatalf("write broken article: %v", err)
	}

	_, err := repo.GetArticles()
	if err == nil {
		t.Fatal("GetArticles() error = nil, want invalid JSON error")
	}
}

func TestGetArticle(t *testing.T) {
	repo := newTestRepository(t)
	article := repository.Article{
		ID:        44,
		CreatedAt: time.Date(2026, 7, 2, 18, 25, 12, 0, time.UTC),
		Title:     "Hello World",
		Text:      "Hello!",
	}
	writeArticle(t, repo.DbPath, article)

	got, err := repo.GetArticle(article.ID)
	if err != nil {
		t.Fatalf("GetArticle(%d) error = %v", article.ID, err)
	}

	if got != article {
		t.Fatalf("GetArticle(%d) = %v, want %v", article.ID, got, article)
	}
}

func TestGetArticleReturnsNotFound(t *testing.T) {
	repo := newTestRepository(t)

	_, err := repo.GetArticle(404)
	if !errors.Is(err, repository.ErrArticleNotFound) {
		t.Fatalf("GetArticle(404) error = %v, want %v", err, repository.ErrArticleNotFound)
	}
}

func TestGetArticleReturnsInvalidJSONError(t *testing.T) {
	repo := newTestRepository(t)
	if err := os.WriteFile(filepath.Join(repo.DbPath, "55.json"), []byte(`{"id":55,"title":`), 0644); err != nil {
		t.Fatalf("write broken article: %v", err)
	}

	_, err := repo.GetArticle(55)
	if err == nil {
		t.Fatal("GetArticle(55) error = nil, want invalid JSON error")
	}
}

func TestAddArticle(t *testing.T) {
	repo := newTestRepository(t)

	id, err := repo.AddArticle("New title", "New body")
	if err != nil {
		t.Fatalf("AddArticle() error = %v", err)
	}

	if id != 1 {
		t.Fatalf("AddArticle() id = %d, want 1", id)
	}

	got, err := repo.GetArticle(id)
	if err != nil {
		t.Fatalf("GetArticle(%d) after AddArticle error = %v", id, err)
	}

	if got.ID != id || got.Title != "New title" || got.Text != "New body" {
		t.Fatalf("added article = %v, want id/title/text to match input", got)
	}

	counter, err := os.ReadFile(filepath.Join(repo.DbPath, "counter.txt"))
	if err != nil {
		t.Fatalf("read counter: %v", err)
	}

	if string(counter) != "1" {
		t.Fatalf("counter = %q, want %q", string(counter), "1")
	}
}

func TestAddArticleReturnsCounterError(t *testing.T) {
	repo := repository.ArticlesMemoryRepository{DbPath: t.TempDir()}

	_, err := repo.AddArticle("Title", "Body")
	if err == nil {
		t.Fatal("AddArticle() error = nil, want counter read error")
	}
}

func TestDeleteArticle(t *testing.T) {
	repo := newTestRepository(t)
	article := repository.Article{ID: 10, Title: "Title", Text: "Body"}
	writeArticle(t, repo.DbPath, article)

	if err := repo.DeleteArticle(article.ID); err != nil {
		t.Fatalf("DeleteArticle(%d) error = %v", article.ID, err)
	}

	_, err := repo.GetArticle(article.ID)
	if !errors.Is(err, repository.ErrArticleNotFound) {
		t.Fatalf("GetArticle(%d) after delete error = %v, want %v", article.ID, err, repository.ErrArticleNotFound)
	}
}

func TestDeleteArticleReturnsNotFound(t *testing.T) {
	repo := newTestRepository(t)

	err := repo.DeleteArticle(404)
	if !errors.Is(err, repository.ErrArticleNotFound) {
		t.Fatalf("DeleteArticle(404) error = %v, want %v", err, repository.ErrArticleNotFound)
	}
}

func TestUpdateArticle(t *testing.T) {
	repo := newTestRepository(t)
	article := repository.Article{ID: 15, Title: "Old title", Text: "Old body"}
	writeArticle(t, repo.DbPath, article)

	if err := repo.UpdateArticle(article.ID, "Updated title", "Updated body"); err != nil {
		t.Fatalf("UpdateArticle(%d) error = %v", article.ID, err)
	}

	got, err := repo.GetArticle(article.ID)
	if err != nil {
		t.Fatalf("GetArticle(%d) after UpdateArticle error = %v", article.ID, err)
	}

	if got.ID != article.ID || got.Title != "Updated title" || got.Text != "Updated body" {
		t.Fatalf("updated article = %v, want updated title/body with same id", got)
	}

	if got.CreatedAt.IsZero() {
		t.Fatal("updated article CreatedAt is zero, want current timestamp")
	}
}

func TestUpdateArticleReturnsNotFound(t *testing.T) {
	repo := newTestRepository(t)

	err := repo.UpdateArticle(404, "Title", "Body")
	if !errors.Is(err, repository.ErrArticleNotFound) {
		t.Fatalf("UpdateArticle(404) error = %v, want %v", err, repository.ErrArticleNotFound)
	}
}

package test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/AndreyTishchenko/Go_projects/personal_blog/repository"
)

func TestGetArticles(t *testing.T) {
	dbPath := t.TempDir()
	article := repository.Article{
		ID:        44,
		CreatedAt: time.Date(2026, 07, 02, 18, 25, 12, 12, *&time.Local),
		Title:     "Hello World",
		Text:      "Hello!",
	}

	json_article, err := json.Marshal(article)

	if err != nil {
		return
	}

	err = os.WriteFile(filepath.Join(dbPath, strconv.Itoa(article.ID)+".json"), json_article, 0644)
	err = os.WriteFile(filepath.Join(dbPath, strconv.Itoa(55)+".json"), []byte(`{"id":44,"title":"Hello World",`), 0644)

	articleMemoryRepository := repository.ArticlesMemoryRepository{
		DbPath: dbPath,
	}

	tests := []struct {
		id      int
		errText string
		result  repository.Article
	}{
		{44, "", article},
		{10, repository.ErrArticleNotFound.Error(), repository.Article{}},
		{55, "unexpected end of JSON input", repository.Article{}},
	}

	for _, tt := range tests {
		article, err := articleMemoryRepository.GetArticle(tt.id)

		if err != nil && err.Error() != tt.errText {
			t.Errorf("GetArticle(%v) expected err be %v got %v", tt.id, tt.errText, err)
		}

		if err == nil && tt.errText != "" {
			t.Errorf("GetArticle(%v) expected err be %v got nil", tt.id, tt.errText)
		}

		if article != tt.result {
			t.Errorf("GetArticle(%v) expected article value be %v got %v", tt.id, tt.result, article)
		}
	}
}

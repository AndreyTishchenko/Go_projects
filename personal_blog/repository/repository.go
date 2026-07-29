package repository

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Article struct {
	ID        int       `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Title     string    `json:"title"`
	Text      string    `json:"text"`
}

type ArticlesRepository interface {
	GetArticles() ([]Article, error)
	GetArticle(id int) (Article, error)
	AddArticle(title string, body string) (int, error)
	DeleteArticle(id int) error
	UpdateArticle(id int, title string, body string) error
	generateID() (int, error)
}

type ArticlesMemoryRepository struct {
	DbPath string
}

var ErrArticleNotFound = errors.New("article not found")

func (r *ArticlesMemoryRepository) GetArticles() ([]Article, error) {
	articles := []Article{}
	err := filepath.WalkDir(r.DbPath, func(path string, d fs.DirEntry, err2 error) error {
		if err2 != nil {
			return err2
		}

		if !d.IsDir() && strings.HasSuffix(d.Name(), ".json") {
			fileData, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			var article Article
			if err := json.Unmarshal(fileData, &article); err != nil {
				return err
			}

			articles = append(articles, article)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return articles, nil
}

func (r *ArticlesMemoryRepository) GetArticle(id int) (Article, error) {
	var article Article

	path := filepath.Join(r.DbPath, strconv.Itoa(id)+".json")

	article_json_data, err := os.ReadFile(path)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Article{}, ErrArticleNotFound
		}
		return Article{}, err
	}

	err = json.Unmarshal(article_json_data, &article)

	if err != nil {
		return Article{}, err
	}

	return article, nil
}

func (r *ArticlesMemoryRepository) AddArticle(title string, body string) (int, error) {
	now := time.Now()

	id, err := r.generateID()

	article := Article{id, now, title, body}

	jsonArticle, err := json.Marshal(article)

	if err != nil {
		return 0, err
	}

	path := filepath.Join(r.DbPath, strconv.Itoa(id)+".json")

	if err := os.WriteFile(path, jsonArticle, 0644); err != nil {
		return 0, err
	}

	return id, nil
}

func (r *ArticlesMemoryRepository) DeleteArticle(id int) error {
	path := filepath.Join(r.DbPath, strconv.Itoa(id)+".json")

	err := os.Remove(path)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ErrArticleNotFound
		}
		return err
	}

	return nil
}

func (r *ArticlesMemoryRepository) UpdateArticle(id int, title string, body string) error {
	path := filepath.Join(r.DbPath, strconv.Itoa(id)+".json")

	now := time.Now()

	article := Article{
		ID:        id,
		CreatedAt: now,
		Title:     title,
		Text:      body,
	}

	jsonArticle, err := json.Marshal(article)

	if err != nil {
		return err
	}

	// err = os.WriteFile(path, jsonArticle, 0644)
	file, err := os.OpenFile(path, os.O_WRONLY, 0644)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ErrArticleNotFound
		}
		return err
	}

	_, err = file.Write(jsonArticle)
	if err != nil {
		return err
	}

	return nil
}

func (r *ArticlesMemoryRepository) generateID() (int, error) {
	path := filepath.Join(r.DbPath, "counter.txt")

	content, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}

	myID, err := strconv.Atoi(string(content))
	if err != nil {
		return 0, err
	}

	myID++

	err = os.WriteFile(path, []byte(strconv.Itoa(myID)), 0644)
	if err != nil {
		return 0, err
	}

	return myID, nil
}

package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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
}

var ErrArticleNotFound = errors.New("article not found")

type ArticlesPostgresRepository struct {
	db *pgxpool.Pool
}

func NewArticlesPostgresRepository(db *pgxpool.Pool) *ArticlesPostgresRepository {
	return &ArticlesPostgresRepository{
		db: db,
	}
}

func (r *ArticlesPostgresRepository) GetArticles() ([]Article, error) {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)

	defer cancel()

	rows, err := r.db.Query(
		ctx,
		`
		SELECT id, created_at, title, text
		FROM articles
		ORDER BY created_at DESC
		`,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	articles := make([]Article, 0)

	for rows.Next() {
		var article Article

		err := rows.Scan(
			&article.ID,
			&article.CreatedAt,
			&article.Title,
			&article.Text,
		)
		if err != nil {
			return nil, err
		}

		articles = append(articles, article)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return articles, nil
}

func (r *ArticlesPostgresRepository) GetArticle(id int) (Article, error) {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	var article Article

	err := r.db.QueryRow(
		ctx,
		`
		SELECT id, created_at, title, text
		FROM articles
		WHERE id = $1
		`,
		id,
	).Scan(
		&article.ID,
		&article.CreatedAt,
		&article.Title,
		&article.Text,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return Article{}, ErrArticleNotFound
	}

	if err != nil {
		return Article{}, err
	}

	return article, nil
}

func (r *ArticlesPostgresRepository) AddArticle(title string, body string) (int, error) {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	var id int

	err := r.db.QueryRow(
		ctx,
		`
		INSERT INTO articles (title, text)
		VALUES ($1, $2)
		RETURNING id
		`,
		title,
		body,
	).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *ArticlesPostgresRepository) DeleteArticle(id int) error {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	result, err := r.db.Exec(
		ctx,
		`
		DELETE FROM articles
		WHERE id = $1
		`,
		id,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrArticleNotFound
	}

	return nil
}

func (r *ArticlesPostgresRepository) UpdateArticle(id int, title string, body string) error {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	result, err := r.db.Exec(
		ctx,
		`
		UPDATE articles
		SET title = $1,
		    text = $2
		WHERE id = $3
		`,
		title,
		body,
		id,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrArticleNotFound
	}

	return nil
}

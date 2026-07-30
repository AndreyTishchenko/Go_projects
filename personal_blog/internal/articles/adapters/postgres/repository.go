package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/AndreyTishchenko/Go_projects/personal_blog/internal/articles/app"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db      *pgxpool.Pool
	timeout time.Duration
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db, timeout: 5 * time.Second}
}

func (r *Repository) List(ctx context.Context) ([]app.Article, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	rows, err := r.db.Query(ctx, `
		SELECT id, created_at, title, text
		FROM articles
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	articles := make([]app.Article, 0)
	for rows.Next() {
		var article app.Article
		if err := rows.Scan(&article.ID, &article.CreatedAt, &article.Title, &article.Text); err != nil {
			return nil, err
		}
		articles = append(articles, article)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return articles, nil
}

func (r *Repository) Get(ctx context.Context, id int) (app.Article, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var article app.Article
	err := r.db.QueryRow(ctx, `
		SELECT id, created_at, title, text
		FROM articles
		WHERE id = $1
	`, id).Scan(&article.ID, &article.CreatedAt, &article.Title, &article.Text)
	if errors.Is(err, pgx.ErrNoRows) {
		return app.Article{}, app.ErrNotFound
	}
	return article, err
}

func (r *Repository) Create(ctx context.Context, title, body string) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var id int
	err := r.db.QueryRow(ctx, `
		INSERT INTO articles (title, text)
		VALUES ($1, $2)
		RETURNING id
	`, title, body).Scan(&id)
	return id, err
}

func (r *Repository) Update(ctx context.Context, id int, title, body string) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	result, err := r.db.Exec(ctx, `
		UPDATE articles
		SET title = $1, text = $2
		WHERE id = $3
	`, title, body, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return app.ErrNotFound
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, id int) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	result, err := r.db.Exec(ctx, `DELETE FROM articles WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return app.ErrNotFound
	}
	return nil
}

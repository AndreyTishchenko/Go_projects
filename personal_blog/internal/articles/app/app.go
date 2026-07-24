package app

import (
	"context"
	"errors"
	"strings"
	"time"
)

var (
	ErrNotFound      = errors.New("article not found")
	ErrTitleRequired = errors.New("article title is required")
	ErrBodyRequired  = errors.New("article body is required")
)

type Article struct {
	ID        int       `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Title     string    `json:"title"`
	Text      string    `json:"text"`
}

type Repository interface {
	List(context.Context) ([]Article, error)
	Get(context.Context, int) (Article, error)
	Create(context.Context, string, string) (int, error)
	Update(context.Context, int, string, string) error
	Delete(context.Context, int) error
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) List(ctx context.Context) ([]Article, error) {
	return s.repository.List(ctx)
}

func (s *Service) Get(ctx context.Context, id int) (Article, error) {
	return s.repository.Get(ctx, id)
}

func (s *Service) Create(ctx context.Context, title, body string) (int, error) {
	if err := validate(title, body); err != nil {
		return 0, err
	}
	return s.repository.Create(ctx, title, body)
}

func (s *Service) Update(ctx context.Context, id int, title, body string) error {
	if err := validate(title, body); err != nil {
		return err
	}
	return s.repository.Update(ctx, id, title, body)
}

func (s *Service) Delete(ctx context.Context, id int) error {
	return s.repository.Delete(ctx, id)
}

func validate(title, body string) error {
	if strings.TrimSpace(title) == "" {
		return ErrTitleRequired
	}
	if strings.TrimSpace(body) == "" {
		return ErrBodyRequired
	}
	return nil
}

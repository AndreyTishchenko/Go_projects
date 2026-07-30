package app

import (
	"context"
	"errors"
	"testing"
)

type repositoryStub struct {
	createdTitle string
	createdBody  string
	createID     int
}

func (r *repositoryStub) List(context.Context) ([]Article, error) {
	return nil, nil
}

func (r *repositoryStub) Get(context.Context, int) (Article, error) {
	return Article{}, ErrNotFound
}

func (r *repositoryStub) Create(_ context.Context, title, body string) (int, error) {
	r.createdTitle = title
	r.createdBody = body
	return r.createID, nil
}

func (r *repositoryStub) Update(context.Context, int, string, string) error {
	return nil
}

func (r *repositoryStub) Delete(context.Context, int) error {
	return nil
}

func TestCreateValidatesArticleWithoutHTTP(t *testing.T) {
	repository := &repositoryStub{createID: 42}
	service := NewService(repository)

	if _, err := service.Create(t.Context(), "", "body"); !errors.Is(err, ErrTitleRequired) {
		t.Fatalf("Create() error = %v, want %v", err, ErrTitleRequired)
	}
	if repository.createdBody != "" {
		t.Fatal("Create() called repository for invalid input")
	}

	id, err := service.Create(t.Context(), "Title", "Body")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if id != 42 || repository.createdTitle != "Title" || repository.createdBody != "Body" {
		t.Fatalf("Create() result = (%d, %q, %q), want (42, Title, Body)", id, repository.createdTitle, repository.createdBody)
	}
}

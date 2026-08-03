package test

import (
	"errors"
	"html/template"
	"testing"

	"github.com/AndreyTishchenko/Go_projects/personal_blog/repository"
	"github.com/AndreyTishchenko/Go_projects/personal_blog/server"
)

type failingRandomSource struct {
	err error
}

func (r failingRandomSource) Read([]byte) (int, error) {
	return 0, r.err
}

func TestNewSessionTokenReturnsRandomReadError(t *testing.T) {
	wantErr := errors.New("random source failed")

	token, err := server.NewSessionToken(failingRandomSource{err: wantErr})

	if !errors.Is(err, wantErr) {
		t.Fatalf("NewSessionToken() error = %v, want %v", err, wantErr)
	}
	if token != "" {
		t.Fatalf("NewSessionToken() token = %q, want empty token after random read failure", token)
	}
}

func TestNewServerConfigStopsWhenSessionTokenCreationFails(t *testing.T) {
	wantErr := errors.New("random source failed")

	_, err := server.NewServerConfig(
		&repository.ArticlesPostgresRepository{},
		template.New("test"),
		failingRandomSource{err: wantErr},
	)

	if !errors.Is(err, wantErr) {
		t.Fatalf("NewServerConfig() error = %v, want %v", err, wantErr)
	}
}

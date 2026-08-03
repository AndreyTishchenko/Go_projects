package server

import (
	"errors"
	"html/template"
	"io"

	"github.com/AndreyTishchenko/Go_projects/personal_blog/repository"
)

var ErrEmptyPasswordField, ErrEmptyNameField, ErrBothFieldsEmpty, ErrBadCredentials, ErrEmptyTitleField, ErrEmptyBodyField, ErrInternalServerError, ErrInvalidFormat = errors.New("Empty Password Field"), errors.New("Empty Name Field"), errors.New("Empty Both fields"), errors.New("Bad credentials"), errors.New("Empty Title Field"), errors.New("Empty Body Field"), errors.New("Internal Server Error"), errors.New("Invalid Format")

func NewSessionToken(randomSource io.Reader) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	b := make([]byte, 32)
	n, err := randomSource.Read(b)
	if err != nil {
		return "", err
	}
	if n != len(b) {
		return "", io.ErrUnexpectedEOF
	}

	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}

	return string(b), nil
}

type Server struct {
	ArticlesRepository repository.ArticlesRepository
	Templates          *template.Template
	adminkey           string
}

func NewServerConfig(r repository.ArticlesRepository, t *template.Template, randomSource io.Reader) (Server, error) {
	token, err := NewSessionToken(randomSource)
	if err != nil {
		return Server{}, err
	}

	return Server{
		ArticlesRepository: r,
		Templates:          t,
		adminkey:           token,
	}, nil
}

type ArticlePayload struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

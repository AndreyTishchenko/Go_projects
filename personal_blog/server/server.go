package server

import (
	"crypto/rand"
	"errors"
	"html/template"

	"github.com/AndreyTishchenko/Go_projects/personal_blog/repository"
)

var ErrEmptyPasswordField, ErrEmptyNameField, ErrBothFieldsEmpty, ErrBadCredentials, ErrEmptyTitleField, ErrEmptyBodyField, ErrInternalServerError, ErrInvalidFormat = errors.New("Empty Password Field"), errors.New("Empty Name Field"), errors.New("Empty Both fields"), errors.New("Bad credentials"), errors.New("Empty Title Field"), errors.New("Empty Body Field"), errors.New("Internal Server Error"), errors.New("Invalid Format")

func newSessionToken() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	b := make([]byte, 32)
	rand.Read(b)

	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}

	return string(b)
}

type Server struct {
	ArticlesRepository repository.ArticlesRepository
	Templates          *template.Template
	adminkey           string
	adminLogin         string
	adminPassword      string
}

func NewServerConfig(r repository.ArticlesRepository, t *template.Template, adminLogin, adminPassword string) Server {
	return Server{
		ArticlesRepository: r,
		Templates:          t,
		adminkey:           newSessionToken(),
		adminLogin:         adminLogin,
		adminPassword:      adminPassword,
	}
}

type ArticlePayload struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

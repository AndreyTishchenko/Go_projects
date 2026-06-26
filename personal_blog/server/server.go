package server

import (
	"errors"
	"html/template"

	"github.com/AndreyTishchenko/Go_projects/personal_blog/repository"
)

var ErrEmptyPasswordField, ErrEmptyNameField, ErrBothFieldsEmpty, ErrBadCredentials, ErrEmptyTitleField, ErrEmptyBodyField, ErrInternalServerError, ErrInvalidFormat = errors.New("Empty Password Field"), errors.New("Empty Name Field"), errors.New("Empty Both fields"), errors.New("Bad credentials"), errors.New("Empty Title Field"), errors.New("Empty Body Field"), errors.New("Internal Server Error"), errors.New("Invalid Format")

type Server struct {
	ArticlesRepository repository.ArticlesRepository
	Templates          *template.Template
	adminkey           string
}

func NewServerConfig(r repository.ArticlesRepository, t *template.Template) Server {
	return Server{
		ArticlesRepository: r,
		Templates:          t,
		adminkey:           "3282h7scc9dh932n9fsndn23noxc",
	}
}

type ArticlePayload struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

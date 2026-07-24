package http

import (
	"errors"
	"html/template"
	"log/slog"

	"github.com/AndreyTishchenko/Go_projects/personal_blog/internal/articles/app"
	"github.com/AndreyTishchenko/Go_projects/personal_blog/internal/platform/session"
)

var (
	ErrEmptyPasswordField = errors.New("Empty Password Field")
	ErrEmptyNameField     = errors.New("Empty Name Field")
	ErrBothFieldsEmpty    = errors.New("Empty Both fields")
	ErrBadCredentials     = errors.New("Bad credentials")
	ErrEmptyTitleField    = errors.New("Empty Title Field")
	ErrEmptyBodyField     = errors.New("Empty Body Field")
	ErrInternalServer     = errors.New("Internal Server Error")
	ErrInvalidFormat      = errors.New("Invalid Format")
)

type Handler struct {
	articles  *app.Service
	templates *template.Template
	sessions  *session.Manager
	logger    *slog.Logger
}

func NewHandler(
	articles *app.Service,
	templates *template.Template,
	sessions *session.Manager,
	logger *slog.Logger,
) *Handler {
	return &Handler{
		articles:  articles,
		templates: templates,
		sessions:  sessions,
		logger:    logger,
	}
}

type articlePayload struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

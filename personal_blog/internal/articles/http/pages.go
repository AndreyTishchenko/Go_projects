package http

import (
	"bytes"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/AndreyTishchenko/Go_projects/personal_blog/internal/articles/app"
	"github.com/go-chi/chi/v5"
)

type ReadibleArticle struct {
	ID        int    `json:"id"`
	CreatedAt string `json:"created_at"`
	Title     string `json:"title"`
	Text      string `json:"text"`
}

func ToReadableArticles(src []app.Article) []ReadibleArticle {
	out := make([]ReadibleArticle, 0, len(src))

	for _, a := range src {
		out = append(out, ReadibleArticle{
			ID:        a.ID,
			CreatedAt: a.CreatedAt.Format("January 2, 2006"),
			Title:     a.Title,
			Text:      a.Text,
		})
	}

	return out
}

func (s *Handler) RenderTemplate(w http.ResponseWriter, status int, name string, data any) {
	var buf bytes.Buffer

	err := s.templates.ExecuteTemplate(&buf, name, data)
	if err != nil {
		http.Error(w, "template rendering failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(status)
	buf.WriteTo(w)
}

type homePageData struct {
	Articles []ReadibleArticle
	IsAdmin  bool
}

func (s *Handler) HomePage(w http.ResponseWriter, r *http.Request) {
	var isAdmin bool
	articles, err := s.articles.List(r.Context())

	if err != nil {
		s.logger.Error("failed to load articles", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	cookie, err := r.Cookie("auth")

	if err != nil {
		if err != http.ErrNoCookie {
			s.logger.Error("cookie parse error", "error", err)
		}
		isAdmin = false
	} else {
		isAdmin = s.sessions.IsAdmin(cookie.Value)
	}

	data := homePageData{ToReadableArticles(articles), isAdmin}

	s.RenderTemplate(w, http.StatusOK, "home.html", data)
}

type articlePageData struct {
	Article ReadibleArticle
}

func (s *Handler) ArticlePage(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)

	if err != nil {
		s.RenderTemplate(w, http.StatusNotFound, "404.html", nil)
		return
	}

	article, err := s.articles.Get(r.Context(), id)

	if err != nil {
		s.RenderTemplate(w, http.StatusNotFound, "404.html", nil)
		return
	}

	data := articlePageData{
		Article: ReadibleArticle{
			ID:        article.ID,
			CreatedAt: article.CreatedAt.Format("January 2, 2006"),
			Title:     article.Title,
			Text:      article.Text,
		},
	}

	s.RenderTemplate(w, http.StatusOK, "article.html", data)
}

type LoginPageData struct {
	NameErr           bool
	PasswordErr       bool
	BadCredentialsErr bool
}

func (s *Handler) LoginPage(w http.ResponseWriter, r *http.Request) {
	errorMsg := getFlash(w, r)
	nameError := false
	passwordError := false
	badCredentialsError := false

	if errorMsg == ErrBothFieldsEmpty.Error() || errorMsg == ErrEmptyNameField.Error() {
		nameError = true
	}

	if errorMsg == ErrBothFieldsEmpty.Error() || errorMsg == ErrEmptyPasswordField.Error() {
		passwordError = true
	}

	if errorMsg == ErrBadCredentials.Error() {
		badCredentialsError = true
	}

	s.RenderTemplate(w, http.StatusOK, "auth.html", LoginPageData{
		nameError,
		passwordError,
		badCredentialsError,
	})
}

type AdminPageData struct {
	Articles []ReadibleArticle
	IsAdmin  bool
}

func (s *Handler) AdminPage(w http.ResponseWriter, r *http.Request) {
	var isAdmin bool
	articles, err := s.articles.List(r.Context())

	if err != nil {
		s.logger.Error("failed to load articles", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	cookie, err := r.Cookie("auth")

	if err != nil {
		if err != http.ErrNoCookie {
			s.logger.Error("cookie parse error", "error", err)
		}
		isAdmin = false
	} else {
		isAdmin = s.sessions.IsAdmin(cookie.Value)
	}

	if isAdmin != true {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	data := AdminPageData{ToReadableArticles(articles), isAdmin}

	s.RenderTemplate(w, http.StatusOK, "admin_panel.html", data)
}

type AddArticleData struct {
	Date      string
	TitleErr  bool
	BodyErr   bool
	TitleText string
	BodyText  string
}

func (s *Handler) AddArticle(w http.ResponseWriter, r *http.Request) {
	var isAdmin bool
	title_text := ""
	body_text := ""

	cookie, err := r.Cookie("auth")

	if err != nil {
		if err != http.ErrNoCookie {
			s.logger.Error("cookie parse error", "error", err)
		}
		isAdmin = false
	} else {
		isAdmin = s.sessions.IsAdmin(cookie.Value)
	}

	if isAdmin != true {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	title_text_cookie, err := r.Cookie("Title")
	if err == nil {
		title_text_text, err := url.QueryUnescape(title_text_cookie.Value)
		if err == nil {
			title_text = title_text_text
		}
	}
	body_text_cookie, err := r.Cookie("Body")
	if err == nil {
		body_text_text, err := url.QueryUnescape(body_text_cookie.Value)
		if err == nil {
			body_text = body_text_text
		}
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "Title",
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // seconds
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "Body",
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // seconds
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	errorMsg := getFlash(w, r)

	titleError := false
	bodyError := false

	if errorMsg == ErrBothFieldsEmpty.Error() || errorMsg == ErrEmptyTitleField.Error() {
		titleError = true
	}

	if errorMsg == ErrBothFieldsEmpty.Error() || errorMsg == ErrEmptyBodyField.Error() {
		bodyError = true
	}

	data := AddArticleData{
		time.Now().Format("January 2, 2006"),
		titleError,
		bodyError,
		title_text,
		body_text,
	}

	s.RenderTemplate(w, http.StatusOK, "add_article_form.html", data)
}

type ChangeArticleData struct {
	Date      string
	TitleErr  bool
	BodyErr   bool
	TitleText string
	BodyText  string
	ID        int
}

func (s *Handler) ChangeArticle(w http.ResponseWriter, r *http.Request) {
	var isAdmin bool
	title_text := ""
	body_text := ""

	cookie, err := r.Cookie("auth")

	if err != nil {
		if err != http.ErrNoCookie {
			s.logger.Error("cookie parse error", "error", err)
		}
		isAdmin = false
	} else {
		isAdmin = s.sessions.IsAdmin(cookie.Value)
	}

	if isAdmin != true {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid article id", http.StatusBadRequest)
		return
	}

	article, err := s.articles.Get(r.Context(), id)
	if err != nil {
		http.Error(w, "article not found", http.StatusNotFound)
		return
	}

	title_text = article.Title
	body_text = article.Text

	title_text_cookie, err := r.Cookie("Title")
	if err == nil {
		title_text_text, err := url.QueryUnescape(title_text_cookie.Value)
		if err == nil {
			title_text = title_text_text
		}
	}

	body_text_cookie, err := r.Cookie("Body")
	if err == nil {
		body_text_text, err := url.QueryUnescape(body_text_cookie.Value)
		if err == nil {
			body_text = body_text_text
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "Title",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "Body",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	errorMsg := getFlash(w, r)

	titleError := false
	bodyError := false

	if errorMsg == ErrBothFieldsEmpty.Error() || errorMsg == ErrEmptyTitleField.Error() {
		titleError = true
	}

	if errorMsg == ErrBothFieldsEmpty.Error() || errorMsg == ErrEmptyBodyField.Error() {
		bodyError = true
	}

	data := ChangeArticleData{
		time.Now().Format("January 2, 2006"),
		titleError,
		bodyError,
		title_text,
		body_text,
		id,
	}

	s.RenderTemplate(w, http.StatusOK, "change_article_form.html", data)
}

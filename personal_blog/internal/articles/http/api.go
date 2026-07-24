package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/AndreyTishchenko/Go_projects/personal_blog/internal/articles/app"
	"github.com/go-chi/chi/v5"
)

func authCheck(w http.ResponseWriter, r *http.Request, s *Handler, isBrowser bool) error {
	cookie, err := r.Cookie("auth")

	if err != nil {
		if err != http.ErrNoCookie {
			s.logger.Error("cookie parse error", "error", err)
			if isBrowser {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return err
			} else {
				http.Error(w, "Cannot read cookies", http.StatusInternalServerError)
				return err
			}
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return err
	} else {
		if s.sessions.IsAdmin(cookie.Value) {
			return nil
		} else {
			if isBrowser {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return errors.New("Not logged In")
			} else {
				http.Error(w, "Must be logged in", http.StatusForbidden)
				return errors.New("Not logged In")
			}
		}
	}
}

func (s *Handler) GetArticles(w http.ResponseWriter, r *http.Request) {
	articles, err := s.articles.List(r.Context())

	if err != nil {
		s.logger.Error("failed to get articles", "error", err)
		http.Error(w, "Failed to read payload", http.StatusInternalServerError)
		return
	}

	jsonArticles, err := json.Marshal(articles)

	if err != nil {
		s.logger.Error("failed to encode articles", "error", err)
		http.Error(w, "Failed to get articles", http.StatusInternalServerError)
		return
	}

	w.Write(jsonArticles)
}

func (s *Handler) GetArticle(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)

	if err != nil {
		http.Error(w, "invalid article id format", http.StatusBadRequest)
		return
	}

	article, err := s.articles.Get(r.Context(), id)

	if err != nil {
		if errors.Is(err, app.ErrNotFound) {
			println("Article is not found:", err)
			http.Error(w, "Article is not found", http.StatusNotFound)
			return
		}
		println("Cannot get article:", err)
		http.Error(w, "Cannot get article", http.StatusInternalServerError)
		return
	}

	jsonArticle, err := json.Marshal(article)

	if err != nil {
		println("Cannot get article:", err)
		http.Error(w, "Cannot get article", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonArticle)
}

func parseArticleData(w http.ResponseWriter, r *http.Request) (articlePayload, error, bool) {
	var article articlePayload
	isBrowser := !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json")

	if isBrowser {
		if err := r.ParseForm(); err != nil {
			http.Redirect(w, r, "/admin/new", http.StatusSeeOther)
			return articlePayload{}, err, true
		}

		article.Body = r.FormValue("body")
		article.Title = r.FormValue("title")
		return article, nil, true
	} else {
		if r.Method != http.MethodPost {
			http.Error(w, "invalid method", http.StatusMethodNotAllowed)
			return articlePayload{}, errors.New("invalid method"), false
		}

		err := json.NewDecoder(r.Body).Decode(&article)

		if err != nil {
			http.Error(w, "invalid format", http.StatusBadRequest)
			return articlePayload{}, err, false
		}

		return article, nil, false
	}
}

func setDraftArticle(w http.ResponseWriter, title string, body string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "Title",
		Value:    url.QueryEscape(title),
		Path:     "/",
		MaxAge:   60, // seconds
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "Body",
		Value:    url.QueryEscape(body),
		Path:     "/",
		MaxAge:   60, // seconds
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func deleteDraftArticle(w http.ResponseWriter) {
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
}

func inputErrorsParsing(w http.ResponseWriter, r *http.Request, body string, title string, backurl string, isBrowser bool) error {
	if isBrowser {
		if body == "" && title == "" {
			SetFlash(w, ErrBothFieldsEmpty.Error())
			http.Redirect(w, r, backurl, http.StatusSeeOther)
			return ErrBothFieldsEmpty
		} else if body == "" {
			SetFlash(w, ErrEmptyBodyField.Error())
			http.Redirect(w, r, backurl, http.StatusSeeOther)
			return ErrEmptyBodyField
		} else if title == "" {
			SetFlash(w, ErrEmptyTitleField.Error())
			http.Redirect(w, r, backurl, http.StatusSeeOther)
			return ErrEmptyTitleField
		}
	} else {
		if title == "" || body == "" {
			http.Error(w, "title and body are required", http.StatusBadRequest)
			return errors.New("title and body are required")
		}
	}

	return nil
}

func (s *Handler) PostArticle(w http.ResponseWriter, r *http.Request) {
	articleData, err, isBrowser := parseArticleData(w, r)

	if err != nil {
		return
	}

	setDraftArticle(w, articleData.Title, articleData.Body)

	err = inputErrorsParsing(w, r, articleData.Body, articleData.Title, "/admin/new", isBrowser)

	if err != nil {
		return
	}

	err = authCheck(w, r, s, isBrowser)

	if err != nil {
		return
	}

	id, err := s.articles.Create(r.Context(), articleData.Title, articleData.Body)

	if err != nil {
		println("failed to create article", err.Error())
		if isBrowser {
			http.Redirect(w, r, "/admin/new", http.StatusSeeOther)
		} else {
			http.Error(w, "failed to create article", http.StatusInternalServerError)
		}

		return
	}

	deleteDraftArticle(w)

	if isBrowser {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]int{
			"id": id,
		})
	}
}

func (s *Handler) UpdateArticle(w http.ResponseWriter, r *http.Request) {
	articleData, err, isBrowser := parseArticleData(w, r)

	if err != nil {
		return
	}

	setDraftArticle(w, articleData.Title, articleData.Body)

	rawID := chi.URLParam(r, "id")

	id, err := strconv.Atoi(rawID)

	if err != nil {
		if isBrowser {
			http.Redirect(w, r, fmt.Sprintf("/admin/change/%s", rawID), http.StatusSeeOther)
			return
		} else {
			http.Error(w, "invalid article id", http.StatusBadRequest)
		}
		return
	}

	err = inputErrorsParsing(w, r, articleData.Body, articleData.Title, fmt.Sprintf("/admin/change/%s", rawID), isBrowser)

	if err != nil {
		return
	}

	err = authCheck(w, r, s, isBrowser)

	if err != nil {
		return
	}

	err = s.articles.Update(r.Context(), id, articleData.Title, articleData.Body)

	if isBrowser {
		if err != nil {
			println("failed to update article", err.Error())
			http.Redirect(w, r, fmt.Sprintf("/admin/change/%s", rawID), http.StatusSeeOther)
			return
		}

		deleteDraftArticle(w)

		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	} else {
		if err != nil {
			println("failed to update article", err.Error())
			http.Error(w, "failed to update article", http.StatusInternalServerError)
			return
		}

		deleteDraftArticle(w)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]int{
			"id": id,
		})
	}
}

func (s *Handler) DeleteArticle(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)

	if err != nil {
		http.Error(w, "invalid article id format", http.StatusBadRequest)
		return
	}

	err = authCheck(w, r, s, true)

	if err != nil {
		return
	}

	err = s.articles.Delete(r.Context(), id)

	if err != nil {
		if errors.Is(err, app.ErrNotFound) {
			println("Article is not found:", err)
			http.Error(w, "Article is not found", http.StatusNotFound)
			return
		}
		println("Cannot get article:", err)
		http.Error(w, "Cannot get article", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
)

type authPayload struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

func decodeAuth(r *http.Request) (authPayload, error) {
	var p authPayload

	ct := r.Header.Get("Content-Type")

	if strings.HasPrefix(ct, "application/json") {
		err := json.NewDecoder(r.Body).Decode(&p)
		if err != nil {
			return p, ErrInvalidFormat
		}
	} else {
		if err := r.ParseForm(); err != nil {
			return p, ErrInternalServer
		}
		p.Name = r.FormValue("name")
		p.Password = r.FormValue("password")
	}
	return p, nil
}

func (s *Handler) authenticate(r *http.Request) (string, time.Time, error) {
	data, err := decodeAuth(r)
	if err != nil {
		return "", time.Time{}, err
	}

	if data.Name == "" && data.Password == "" {
		return "", time.Time{}, ErrBothFieldsEmpty
	} else if data.Name == "" {
		return "", time.Time{}, ErrEmptyNameField
	} else if data.Password == "" {
		return "", time.Time{}, ErrEmptyPasswordField
	}

	token, expiresAt, ok := s.sessions.Authenticate(data.Name, data.Password)
	if !ok {
		return "", time.Time{}, ErrBadCredentials
	}

	return token, expiresAt, nil
}

func (s *Handler) AuthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
		return
	}

	token, expiresAt, err := s.authenticate(r)
	isBrowser := !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json")

	if err != nil {
		if isBrowser {
			if errors.Is(err, ErrBothFieldsEmpty) {
				SetFlash(w, ErrBothFieldsEmpty.Error())
				http.Redirect(w, r, "/login", http.StatusSeeOther)
			} else if errors.Is(err, ErrEmptyNameField) {
				SetFlash(w, ErrEmptyNameField.Error())
				http.Redirect(w, r, "/login", http.StatusSeeOther)
			} else if errors.Is(err, ErrEmptyPasswordField) {
				SetFlash(w, ErrEmptyPasswordField.Error())
				http.Redirect(w, r, "/login", http.StatusSeeOther)
			} else if errors.Is(err, ErrBadCredentials) {
				SetFlash(w, ErrBadCredentials.Error())
				http.Redirect(w, r, "/login", http.StatusSeeOther)
			}
		} else {
			http.Error(w, "invalid login or password", http.StatusUnauthorized)
		}
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Expires:  expiresAt,
	})

	if isBrowser {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Handler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("auth"); err == nil {
		s.sessions.Logout(cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(1, 0),
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

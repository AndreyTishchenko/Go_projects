package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type authPayload struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

func (s Server) sign() string {
	return s.adminkey
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
			return p, ErrInternalServerError
		}
		p.Name = r.FormValue("name")
		p.Password = r.FormValue("password")
	}
	return p, nil
}

func (s Server) authenticate(r *http.Request) (string, error) {
	data, err := decodeAuth(r)
	if err != nil {
		return "", err
	}

	if data.Name == "" && data.Password == "" {
		return "", ErrBothFieldsEmpty
	} else if data.Name == "" {
		return "", ErrEmptyNameField
	} else if data.Password == "" {
		return "", ErrEmptyPasswordField
	}

	if data.Password != s.adminPassword || data.Name != s.adminLogin {
		return "", ErrBadCredentials
	}

	return s.sign(), nil
}

func (s Server) AuthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
		return
	}

	token, err := s.authenticate(r)
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
	})

	if isBrowser {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s Server) AuthCheck(key string) bool {
	if key == s.adminkey {
		return true
	}
	return false
}

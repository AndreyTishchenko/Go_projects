package http

import (
	"net/http"
	"net/url"
)

func SetFlash(w http.ResponseWriter, message string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "flash",
		Value:    url.QueryEscape(message),
		Path:     "/",
		MaxAge:   60, // seconds
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func getFlash(w http.ResponseWriter, r *http.Request) string {
	flashData, err := r.Cookie("flash")

	if err != nil {
		return ""
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "flash",
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // seconds
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	data, err := url.QueryUnescape(flashData.Value)

	if err != nil {
		return ""
	}

	return data
}

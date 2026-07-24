package test

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/AndreyTishchenko/Go_projects/personal_blog/internal/articles/app"
)

func TestAdminRoutesRequireAuthAndDeleteArticle(t *testing.T) {
	repo := newFakeArticlesRepository(app.Article{
		ID:        9,
		CreatedAt: time.Date(2026, 7, 2, 18, 25, 12, 0, time.UTC),
		Title:     "Delete me",
		Text:      "Body",
	})
	_, handler := newTestApplication(t, repo)

	rr := performRequest(handler, http.MethodGet, "/admin/", "")
	if rr.Code != http.StatusFound {
		t.Fatalf("GET /admin/ without auth status = %d, want %d", rr.Code, http.StatusFound)
	}
	if location := rr.Header().Get("Location"); location != "/login" {
		t.Fatalf("GET /admin/ without auth Location = %q, want %q", location, "/login")
	}

	authCookie := loginAsAdmin(t, handler)
	rr = performRequest(handler, http.MethodPost, "/admin/delete/9", "", authCookie)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("POST /admin/delete/9 status = %d, want %d; body = %q", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
	if location := rr.Header().Get("Location"); location != "/admin" {
		t.Fatalf("POST /admin/delete/9 Location = %q, want %q", location, "/admin")
	}

	if _, err := repo.Get(t.Context(), 9); !errors.Is(err, app.ErrNotFound) {
		t.Fatalf("Get(9) after delete error = %v, want %v", err, app.ErrNotFound)
	}
}

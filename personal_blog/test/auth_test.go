package test

import (
	"bytes"
	"log/slog"
	"net/http"
	"strings"
	"testing"
)

func TestAuthRejectsBadJSONCredentials(t *testing.T) {
	_, handler := newTestApplication(t, newFakeArticlesRepository())

	rr := performRequest(handler, http.MethodPost, "/auth", `{"name":"admin","password":"wrong"}`)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("POST /auth with bad credentials status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
	if len(rr.Result().Cookies()) != 0 {
		t.Fatalf("POST /auth with bad credentials cookies = %v, want none", rr.Result().Cookies())
	}
}

func TestLogoutClearsAndInvalidatesSession(t *testing.T) {
	_, handler := newTestApplication(t, newFakeArticlesRepository())
	authCookie := loginAsAdmin(t, handler)

	rr := performRequest(handler, http.MethodPost, "/logout", "", authCookie)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("POST /logout status = %d, want %d", rr.Code, http.StatusSeeOther)
	}
	if location := rr.Header().Get("Location"); location != "/login" {
		t.Fatalf("POST /logout Location = %q, want /login", location)
	}

	var cleared bool
	for _, cookie := range rr.Result().Cookies() {
		if cookie.Name == "auth" && cookie.MaxAge < 0 {
			cleared = true
		}
	}
	if !cleared {
		t.Fatal("POST /logout did not clear the auth cookie")
	}

	rr = performRequest(handler, http.MethodGet, "/admin/", "", authCookie)
	if rr.Code != http.StatusFound {
		t.Fatalf("GET /admin/ after logout status = %d, want %d", rr.Code, http.StatusFound)
	}
}

func TestHealthCheck(t *testing.T) {
	_, handler := newTestApplication(t, newFakeArticlesRepository())

	rr := performRequest(handler, http.MethodGet, "/healthz", "")
	if rr.Code != http.StatusOK {
		t.Fatalf("GET /healthz status = %d, want %d", rr.Code, http.StatusOK)
	}
	if rr.Body.String() != "OK\n" {
		t.Fatalf("GET /healthz body = %q, want %q", rr.Body.String(), "OK\n")
	}
}

func TestRequestLogIncludesHTTPDetails(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))
	_, handler := newTestApplicationWithLogger(t, newFakeArticlesRepository(), logger)

	performRequest(handler, http.MethodGet, "/healthz", "")

	logLine := output.String()
	for _, field := range []string{
		`"method":"GET"`,
		`"path":"/healthz"`,
		`"status":200`,
		`"duration":`,
	} {
		if !strings.Contains(logLine, field) {
			t.Errorf("request log %q does not contain %q", logLine, field)
		}
	}
}

package test

import (
	"net/http"
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

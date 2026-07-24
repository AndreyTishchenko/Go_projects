package session

import (
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func TestSessionExpires(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}

	now := time.Date(2026, 7, 24, 12, 0, 0, 0, time.UTC)
	manager := NewManager("admin", string(hash), time.Hour)
	manager.now = func() time.Time { return now }

	token, expiresAt, ok := manager.Authenticate("admin", "secret")
	if !ok {
		t.Fatal("Authenticate() rejected valid credentials")
	}
	if want := now.Add(time.Hour); !expiresAt.Equal(want) {
		t.Fatalf("expiration = %v, want %v", expiresAt, want)
	}
	if !manager.IsAdmin(token) {
		t.Fatal("IsAdmin() rejected an active session")
	}

	now = expiresAt
	if manager.IsAdmin(token) {
		t.Fatal("IsAdmin() accepted an expired session")
	}
}

func TestLogoutInvalidatesSession(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}

	manager := NewManager("admin", string(hash), time.Hour)
	token, _, ok := manager.Authenticate("admin", "secret")
	if !ok {
		t.Fatal("Authenticate() rejected valid credentials")
	}

	manager.Logout(token)
	if manager.IsAdmin(token) {
		t.Fatal("IsAdmin() accepted a logged-out session")
	}
}

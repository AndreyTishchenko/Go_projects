package session

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Manager struct {
	mu                sync.Mutex
	adminLogin        string
	adminPasswordHash []byte
	sessionTTL        time.Duration
	sessions          map[string]time.Time
	now               func() time.Time
}

func NewManager(adminLogin, adminPasswordHash string, sessionTTL time.Duration) *Manager {
	return &Manager{
		adminLogin:        adminLogin,
		adminPasswordHash: []byte(adminPasswordHash),
		sessionTTL:        sessionTTL,
		sessions:          make(map[string]time.Time),
		now:               time.Now,
	}
}

func (m *Manager) Authenticate(login, password string) (string, time.Time, bool) {
	loginOK := subtle.ConstantTimeCompare([]byte(login), []byte(m.adminLogin)) == 1
	passwordOK := bcrypt.CompareHashAndPassword(m.adminPasswordHash, []byte(password)) == nil
	if !loginOK || !passwordOK {
		return "", time.Time{}, false
	}

	token := newToken()
	expiresAt := m.now().Add(m.sessionTTL)

	m.mu.Lock()
	for existingToken, existingExpiry := range m.sessions {
		if !m.now().Before(existingExpiry) {
			delete(m.sessions, existingToken)
		}
	}
	m.sessions[token] = expiresAt
	m.mu.Unlock()

	return token, expiresAt, true
}

func (m *Manager) IsAdmin(token string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	expiresAt, ok := m.sessions[token]
	if !ok {
		return false
	}
	if !m.now().Before(expiresAt) {
		delete(m.sessions, token)
		return false
	}
	return true
}

func (m *Manager) Logout(token string) {
	m.mu.Lock()
	delete(m.sessions, token)
	m.mu.Unlock()
}

func newToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic("generate session token: " + err.Error())
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

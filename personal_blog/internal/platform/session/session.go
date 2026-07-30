package session

import (
	"crypto/rand"
	"crypto/subtle"
)

type Manager struct {
	adminLogin    string
	adminPassword string
	adminToken    string
}

func NewManager(adminLogin, adminPassword string) *Manager {
	return &Manager{
		adminLogin:    adminLogin,
		adminPassword: adminPassword,
		adminToken:    newToken(),
	}
}

func (m *Manager) Authenticate(login, password string) (string, bool) {
	loginOK := subtle.ConstantTimeCompare([]byte(login), []byte(m.adminLogin)) == 1
	passwordOK := subtle.ConstantTimeCompare([]byte(password), []byte(m.adminPassword)) == 1
	return m.adminToken, loginOK && passwordOK
}

func (m *Manager) IsAdmin(token string) bool {
	return subtle.ConstantTimeCompare([]byte(token), []byte(m.adminToken)) == 1
}

func newToken() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic("generate session token: " + err.Error())
	}
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b)
}

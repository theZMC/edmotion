package admin

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"sync"
)

type State struct {
	mu             sync.RWMutex
	password       string
	giveFixedFiles bool
}

func NewState(password string, giveFixedFiles bool) (*State, string, error) {
	if password != "" {
		return &State{password: password, giveFixedFiles: giveFixedFiles}, "", nil
	}

	generatedPassword, err := generatePassword()
	if err != nil {
		return nil, "", fmt.Errorf("generating random password: %w", err)
	}

	state := &State{
		password:       generatedPassword,
		giveFixedFiles: giveFixedFiles,
	}

	return state, generatedPassword, nil
}

func (s *State) IsAuthorized(r *http.Request) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return r.Header.Get("Authorization") == s.password
}

func (s *State) ToggleFixedFiles() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.giveFixedFiles = !s.giveFixedFiles
	return s.giveFixedFiles
}

func (s *State) SetPassword(password string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.password = password
}

func (s *State) ServeFixedFiles() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.giveFixedFiles
}

func generatePassword() (string, error) {
	randBytes := make([]byte, 16)
	if _, err := rand.Read(randBytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(randBytes), nil
}

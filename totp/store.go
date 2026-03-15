package totp

import (
	"sync"
)

type Store struct {
	mu      sync.RWMutex
	secrets map[string]string
}

func NewStore() *Store {
	return &Store{
		secrets: make(map[string]string),
	}
}

func (s *Store) Set(accountName, secret string) (isNew bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.secrets[accountName]
	s.secrets[accountName] = secret
	return !exists
}

func (s *Store) Get(accountName string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	secret, exists := s.secrets[accountName]
	return secret, exists
}

func (s *Store) Delete(accountName string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.secrets[accountName]
	if exists {
		delete(s.secrets, accountName)
	}
	return exists
}

package store

import (
	"errors"
	"golang/models"
	"sync"

	"github.com/google/uuid"
)

var ErrUserExists = errors.New("user already exists")
var ErrUserNotFound = errors.New("user not found")

type MemStore struct {
	mu    sync.RWMutex
	users map[string]*models.User // key by username
}

func NewStore() *MemStore {
	return &MemStore{users: map[string]*models.User{}}
}

func (s *MemStore) CreateUser(username, hashedPassword, role string) (*models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.users[username]; ok {
		return nil, ErrUserExists
	}
	u := &models.User{ID: uuid.NewString(), Username: username, Password: hashedPassword, Role: role}
	s.users[username] = u
	return u, nil
}

func (s *MemStore) GetByUsername(username string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[username]
	if !ok {
		return nil, ErrUserNotFound
	}
	return u, nil
}

package service

import (
	"fmt"

	"example.com/go-api/internal/domain"
)

type UserService struct {
	store map[string]*domain.User
	seq   int
}

func NewUserService() *UserService {
	return &UserService{
		store: make(map[string]*domain.User),
	}
}

func (s *UserService) FindByID(id string) (*domain.User, error) {
	user, ok := s.store[id]
	if !ok {
		return nil, fmt.Errorf("user %s not found", id)
	}
	return user, nil
}

func (s *UserService) Create(name, email string) *domain.User {
	s.seq++
	id := fmt.Sprintf("usr_%d", s.seq)
	user := &domain.User{
		ID:    id,
		Name:  name,
		Email: email,
	}
	s.store[id] = user
	return user
}

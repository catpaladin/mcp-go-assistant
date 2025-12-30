// Package example demonstrates good Go practices
package example

import (
	"context"
	"fmt"
)

// User represents a user in the system
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// UserService handles user operations
type UserService struct {
	repo UserRepository
}

// UserRepository defines the interface for user data access
type UserRepository interface {
	GetUser(ctx context.Context, id int) (*User, error)
	SaveUser(ctx context.Context, user *User) error
}

// NewUserService creates a new UserService instance
func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

// GetUser retrieves a user by ID with proper error handling
func (s *UserService) GetUser(ctx context.Context, id int) (*User, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid user ID: %d", id)
	}

	user, err := s.repo.GetUser(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user %d: %w", id, err)
	}

	return user, nil
}

// FormatUserInfo returns a formatted string representation of a user
func (s *UserService) FormatUserInfo(ctx context.Context, id int) (string, error) {
	user, err := s.GetUser(ctx, id)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("User: %s (%s)", user.Name, user.Email), nil
}

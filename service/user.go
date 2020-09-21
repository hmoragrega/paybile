package service

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrInvalidPassword = errors.New("password incorrect")
)

type UserReader interface {
	// GetByLogin load a user by the login.
	// Errors:
	// - ErrUserNotFound: If no user exists with the given login.
	GetByLogin(ctx context.Context, login string) (User, error)
}

type UserService struct {
	Reader UserReader
}

func (svc *UserService) Login(ctx context.Context, login string, password string) (u User, err error) {
	u, err = svc.Reader.GetByLogin(ctx, login)
	if err != nil {
		return u, fmt.Errorf("cannot login user: %w", err)
	}

	if !u.IsCorrectPass(password) {
		return u, fmt.Errorf("cannot login user: %w", ErrInvalidPassword)
	}

	return u, nil
}

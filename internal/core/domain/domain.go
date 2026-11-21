package core_domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID       uuid.UUID
	Email    string
	PassHash string
	IsAdmin  bool
	IsActive bool
}

type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	FamilyID  uuid.UUID
	TokenHash []byte
	UserAgent string
	ExpiresAt time.Time
}

type Token struct {
	Access    string
	Refresh   string
	ExpiresIn int64 // lifetime
}

var (
	ErrUserExists         = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

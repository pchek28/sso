package auth_service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	core_domain "github.com/pchek28/sso/internal/core/domain"
	core_security "github.com/pchek28/sso/internal/core/security"
)

// Что сервису нужно от хранилища.
type Repository interface {
	CreateUser(ctx context.Context, u core_domain.User) error
	UserByEmail(ctx context.Context, email string) (core_domain.User, error)
	SaveRefreshToken(ctx context.Context, t core_domain.RefreshToken) error
}

type Service struct {
	repo       Repository
	jwt        *core_security.JWTIssuer
	refreshTTL time.Duration
	dummyHash  string // для одинакового времени ответа, когда email не найден
}

func New(repo Repository, jwt *core_security.JWTIssuer, refreshTTL time.Duration) (*Service, error) {
	dummy, err := core_security.HashPassword("dummy-password-for-timing")
	if err != nil {
		return nil, fmt.Errorf("init dummy hash: %w", err)
	}

	return &Service{
		repo:       repo,
		jwt:        jwt,
		refreshTTL: refreshTTL,
		dummyHash:  dummy,
	}, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func (s *Service) Register(ctx context.Context, email, password string) (uuid.UUID, error) {
	hash, err := core_security.HashPassword(password)
	if err != nil {
		return uuid.Nil, fmt.Errorf("hash password: %w", err)
	}

	user := core_domain.User{
		ID:       uuid.New(),
		Email:    normalizeEmail(email),
		PassHash: hash,
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return uuid.Nil, err // ErrUserExists пройдёт наверх как есть
	}

	return user.ID, nil
}

func (s *Service) Login(
	ctx context.Context,
	email, password, userAgent string,
) (core_domain.Tokens, core_domain.User, error) {
	user, err := s.repo.UserByEmail(ctx, normalizeEmail(email))
	if err != nil {
		if errors.Is(err, core_domain.ErrUserNotFound) {
			// Сравниваем с фиктивным хешем, чтобы по времени ответа
			// нельзя было понять, существует ли такой email.
			core_security.CheckPassword(s.dummyHash, password)
			return core_domain.Tokens{}, core_domain.User{}, core_domain.ErrInvalidCredentials
		}
		return core_domain.Tokens{}, core_domain.User{}, fmt.Errorf("find user: %w", err)
	}

	if !core_security.CheckPassword(user.PassHash, password) || !user.IsActive {
		return core_domain.Tokens{}, core_domain.User{}, core_domain.ErrInvalidCredentials
	}

	access, err := s.jwt.NewAccessToken(user.ID, user.IsAdmin)
	if err != nil {
		return core_domain.Tokens{}, core_domain.User{}, fmt.Errorf("issue access token: %w", err)
	}

	rawRefresh, refreshHash, err := core_security.NewRefreshToken()
	if err != nil {
		return core_domain.Tokens{}, core_domain.User{}, fmt.Errorf("issue refresh token: %w", err)
	}

	err = s.repo.SaveRefreshToken(ctx, core_domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		FamilyID:  uuid.New(), // новая цепочка токенов при каждом входе
		TokenHash: refreshHash,
		UserAgent: userAgent,
		ExpiresAt: time.Now().Add(s.refreshTTL),
	})
	if err != nil {
		return core_domain.Tokens{}, core_domain.User{}, fmt.Errorf("save refresh token: %w", err)
	}

	return core_domain.Tokens{
		Access:    access,
		Refresh:   rawRefresh,
		ExpiresIn: int64(s.jwt.TTL().Seconds()),
	}, user, nil
}

package auth_repository

import (
	"context"
	"errors"
	"fmt"

	core_domain "github.com/pchek28/sso/internal/core/domain"
	core_postgres_pool "github.com/pchek28/sso/internal/core/repository/postgres"
)

type Repository struct {
	pool core_postgres_pool.Pool
}

func New(pool core_postgres_pool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) CreateUser(ctx context.Context, u core_domain.User) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeOut())
	defer cancel()

	_, err := r.pool.Exec(ctx,
		`INSERT INTO users (id, email, pass_hash) VALUES ($1, $2, $3)`,
		u.ID, u.Email, u.PassHash,
	)
	if err != nil {
		if errors.Is(err, core_postgres_pool.ErrUniqueViolation) {
			return core_domain.ErrUserExists
		}
		return fmt.Errorf("insert user: %w", err)
	}

	return nil
}

func (r *Repository) UserByEmail(ctx context.Context, email string) (core_domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeOut())
	defer cancel()

	var u core_domain.User

	err := r.pool.QueryRow(ctx,
		`SELECT id, email, pass_hash, is_admin, is_active FROM users WHERE email = $1`,
		email,
	).Scan(&u.ID, &u.Email, &u.PassHash, &u.IsAdmin, &u.IsActive)
	if err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return core_domain.User{}, core_domain.ErrUserNotFound
		}
		return core_domain.User{}, fmt.Errorf("select user by email: %w", err)
	}

	return u, nil
}

func (r *Repository) SaveRefreshToken(ctx context.Context, t core_domain.RefreshToken) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeOut())
	defer cancel()

	_, err := r.pool.Exec(ctx,
		`INSERT INTO refresh_tokens (id, user_id, family_id, token_hash, user_agent, expires_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		t.ID, t.UserID, t.FamilyID, t.TokenHash, t.UserAgent, t.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("insert refresh token: %w", err)
	}

	return nil
}

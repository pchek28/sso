package auth_transport_grpc

import (
	"context"
	"errors"
	"net/mail"
	"strings"

	"github.com/google/uuid"
	ssov1 "github.com/pchek28/protos/gen/go/sso"
	core_domain "github.com/pchek28/sso/internal/core/domain"
	core_logger "github.com/pchek28/sso/internal/core/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Что хендлерам нужно от бизнес-логики. Типы обычные, без proto.
type Service interface {
	Register(ctx context.Context, email, password string) (uuid.UUID, error)
	Login(ctx context.Context, email, password, userAgent string) (core_domain.Tokens, core_domain.User, error)
}

type Handler struct {
	ssov1.UnimplementedAuthServer // обязательно по значению
	svc                           Service
	log                           *core_logger.Logger
}

func New(svc Service, log *core_logger.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

func (h *Handler) Register(ctx context.Context, req *ssov1.RegisterRequest) (*ssov1.RegisterResponse, error) {
	email := strings.TrimSpace(req.GetEmail())
	password := req.GetPassword()

	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		return nil, status.Error(codes.InvalidArgument, "invalid email")
	}
	// bcrypt принимает не больше 72 байт
	if n := len(password); n < 8 || n > 72 {
		return nil, status.Error(codes.InvalidArgument, "password must be 8-72 bytes")
	}

	id, err := h.svc.Register(ctx, email, password)
	if err != nil {
		return nil, h.toStatus(err)
	}

	return &ssov1.RegisterResponse{UserId: id.String()}, nil
}

func (h *Handler) Login(ctx context.Context, req *ssov1.LoginRequest) (*ssov1.LoginResponse, error) {
	if req.GetEmail() == "" || req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password are required")
	}

	tokens, user, err := h.svc.Login(ctx, req.GetEmail(), req.GetPassword(), req.GetUserAgent())
	if err != nil {
		return nil, h.toStatus(err)
	}

	return &ssov1.LoginResponse{
		AccessToken:  tokens.Access,
		RefreshToken: tokens.Refresh,
		ExpiresIn:    tokens.ExpiresIn,
		UserId:       user.ID.String(),
		IsAdmin:      user.IsAdmin,
	}, nil
}

// Доменные ошибки превращаются в gRPC-коды; подробности внутренних ошибок только в лог.
func (h *Handler) toStatus(err error) error {
	switch {
	case errors.Is(err, core_domain.ErrUserExists):
		return status.Error(codes.AlreadyExists, "user already exists")
	case errors.Is(err, core_domain.ErrInvalidCredentials):
		return status.Error(codes.Unauthenticated, "invalid email or password")
	default:
		h.log.Error("internal error", zap.Error(err))
		return status.Error(codes.Internal, "internal error")
	}
}

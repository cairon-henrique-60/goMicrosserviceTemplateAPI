package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/cairon-henrique-60/goMicrosserviceTemplateAPI/apps/auth-service/internal/user"
	platformauth "github.com/cairon-henrique-60/goMicrosserviceTemplateAPI/pkg/platform/auth"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type Tokens struct {
	AccessToken      string    `json:"accessToken"`
	RefreshToken     string    `json:"refreshToken"`
	AccessExpiresAt  time.Time `json:"accessExpiresAt"`
	RefreshExpiresAt time.Time `json:"refreshExpiresAt"`
}
type Service struct {
	users      *user.Repository
	sessions   *SessionStore
	jwt        *platformauth.Manager
	refreshTTL time.Duration
}

func NewService(users *user.Repository, sessions *SessionStore, jwt *platformauth.Manager, refreshTTL time.Duration) *Service {
	return &Service{users, sessions, jwt, refreshTTL}
}

func splitRoles(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if v := strings.TrimSpace(p); v != "" {
			out = append(out, v)
		}
	}
	return out
}

func (s *Service) Register(ctx context.Context, name, email, password string) (*user.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	u := &user.User{Name: strings.TrimSpace(name), Email: strings.ToLower(strings.TrimSpace(email)), PasswordHash: string(hash), Roles: "user", Active: true}
	if err := s.users.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (Tokens, error) {
	u, err := s.users.ByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if err != nil || !u.Active || bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return Tokens{}, ErrInvalidCredentials
	}
	return s.issue(ctx, u)
}

func (s *Service) Refresh(ctx context.Context, raw string) (Tokens, error) {
	claims, err := s.jwt.ParseRefresh(raw)
	if err != nil {
		return Tokens{}, ErrInvalidCredentials
	}
	if err := s.sessions.Validate(ctx, claims.ID, raw); err != nil {
		return Tokens{}, ErrInvalidCredentials
	}
	_ = s.sessions.Delete(ctx, claims.ID)
	u, err := s.users.ByID(ctx, claims.UserID)
	if err != nil || !u.Active {
		return Tokens{}, ErrInvalidCredentials
	}
	return s.issue(ctx, u)
}

func (s *Service) Logout(ctx context.Context, raw string) error {
	claims, err := s.jwt.ParseRefresh(raw)
	if err != nil {
		return nil
	}
	return s.sessions.Delete(ctx, claims.ID)
}

func (s *Service) issue(ctx context.Context, u *user.User) (Tokens, error) {
	roles := splitRoles(u.Roles)
	access, accessExp, err := s.jwt.NewAccessToken(u.ID, u.Email, roles)
	if err != nil {
		return Tokens{}, err
	}
	refresh, refreshExp, err := s.jwt.NewRefreshToken(u.ID, u.Email, roles)
	if err != nil {
		return Tokens{}, err
	}
	claims, err := s.jwt.ParseRefresh(refresh)
	if err != nil {
		return Tokens{}, err
	}
	if err := s.sessions.Save(ctx, claims.ID, refresh, s.refreshTTL); err != nil {
		return Tokens{}, err
	}
	return Tokens{access, refresh, accessExp, refreshExp}, nil
}

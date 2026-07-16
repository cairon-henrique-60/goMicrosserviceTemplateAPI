package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID    string   `json:"uid"`
	Email     string   `json:"email"`
	Roles     []string `json:"roles"`
	TokenType string   `json:"typ"`
	jwt.RegisteredClaims
}

type Manager struct {
	accessSecret  []byte
	refreshSecret []byte
	issuer        string
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

func NewManager(accessSecret, refreshSecret, issuer string, accessTTL, refreshTTL time.Duration) *Manager {
	return &Manager{[]byte(accessSecret), []byte(refreshSecret), issuer, accessTTL, refreshTTL}
}

func (m *Manager) NewAccessToken(userID, email string, roles []string) (string, time.Time, error) {
	return m.sign(m.accessSecret, userID, email, roles, "access", m.accessTTL)
}
func (m *Manager) NewRefreshToken(userID, email string, roles []string) (string, time.Time, error) {
	return m.sign(m.refreshSecret, userID, email, roles, "refresh", m.refreshTTL)
}
func (m *Manager) sign(secret []byte, userID, email string, roles []string, typ string, ttl time.Duration) (string, time.Time, error) {
	now := time.Now().UTC()
	exp := now.Add(ttl)
	claims := Claims{UserID: userID, Email: email, Roles: roles, TokenType: typ, RegisteredClaims: jwt.RegisteredClaims{Issuer: m.issuer, Subject: userID, ID: uuid.NewString(), IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(exp)}}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
	return token, exp, err
}
func (m *Manager) ParseAccess(token string) (*Claims, error) {
	return m.parse(token, m.accessSecret, "access")
}
func (m *Manager) ParseRefresh(token string) (*Claims, error) {
	return m.parse(token, m.refreshSecret, "refresh")
}
func (m *Manager) parse(raw string, secret []byte, expectedType string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return secret, nil
	}, jwt.WithIssuer(m.issuer), jwt.WithExpirationRequired())
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}
	if claims.TokenType != expectedType {
		return nil, errors.New("invalid token type")
	}
	return claims, nil
}

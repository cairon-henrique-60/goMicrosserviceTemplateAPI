package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/redis/go-redis/v9"
	"time"
)

var ErrInvalidSession = errors.New("invalid session")

type SessionStore struct{ client *redis.Client }

func NewSessionStore(client *redis.Client) *SessionStore { return &SessionStore{client} }
func hash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
func (s *SessionStore) Save(ctx context.Context, jti, token string, ttl time.Duration) error {
	return s.client.Set(ctx, "session:"+jti, hash(token), ttl).Err()
}
func (s *SessionStore) Validate(ctx context.Context, jti, token string) error {
	value, err := s.client.Get(ctx, "session:"+jti).Result()
	if err != nil || value != hash(token) {
		return ErrInvalidSession
	}
	return nil
}
func (s *SessionStore) Delete(ctx context.Context, jti string) error {
	return s.client.Del(ctx, "session:"+jti).Err()
}

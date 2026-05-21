package session

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	SessionDuration = 24 * time.Hour
	CookieName      = "fleet_session"
)

type Data struct {
	UserID uuid.UUID `json:"user_id"`
	Role   string    `json:"role"`
}

type Store struct {
	rdb *redis.Client
}

var ErrNotFound = errors.New("session not found")

func New() *Store {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: os.Getenv("REDIS_PASSWORD"),
	})
	return &Store{rdb: rdb}
}

func (s *Store) key(id string) string { return "session:" + id }

func (s *Store) Create(ctx context.Context, data Data) (string, error) {
	id := uuid.NewString()
	b, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	if err := s.rdb.Set(ctx, s.key(id), b, SessionDuration).Err(); err != nil {
		return "", err
	}
	return id, nil
}

func (s *Store) Get(ctx context.Context, id string) (*Data, error) {
	b, err := s.rdb.GetEx(ctx, s.key(id), SessionDuration).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	var d Data
	if err := json.Unmarshal(b, &d); err != nil {
		return nil, err
	}
	return &d, nil
}

func (s *Store) Delete(ctx context.Context, id string) error {
	return s.rdb.Del(ctx, s.key(id)).Err()
}

func (s *Store) DeleteAllForUser(ctx context.Context, userID uuid.UUID) error {
	setKey := s.userSetKey(userID)
	ids, err := s.rdb.SMembers(ctx, setKey).Result()
	if err != nil {
		return err
	}
	pipe := s.rdb.Pipeline()
	for _, id := range ids {
		pipe.Del(ctx, s.key(id))
	}
	pipe.Del(ctx, setKey)
	_, err = pipe.Exec(ctx)
	return err
}

func (s *Store) TrackUserSession(ctx context.Context, userID uuid.UUID, sessionID string) error {
	setKey := s.userSetKey(userID)
	pipe := s.rdb.Pipeline()
	pipe.SAdd(ctx, setKey, sessionID)
	pipe.Expire(ctx, setKey, SessionDuration*2)
	_, err := pipe.Exec(ctx)
	return err
}

func (s *Store) userSetKey(userID uuid.UUID) string {
	return "user_sessions:" + userID.String()
}

func (s *Store) Ping(ctx context.Context) error {
	return s.rdb.Ping(ctx).Err()
}

package store

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/kapibara-pro/sway-project/backend/internal/domain"
)

var ErrNotFound = errors.New("not found")

type Store interface {
	GetProviderConfig(ctx context.Context, deviceID string) (domain.ProviderConfig, error)
	SaveProviderConfig(ctx context.Context, cfg domain.ProviderConfig) error
	DeleteProviderConfig(ctx context.Context, deviceID string) error
	SaveRequestLog(ctx context.Context, log domain.RequestLog) error
	ListRequestLogs(ctx context.Context, deviceID string, limit int) ([]domain.RequestLog, error)
}

type MemoryStore struct {
	mu        sync.RWMutex
	providers map[string]domain.ProviderConfig
	logs      []domain.RequestLog
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{providers: make(map[string]domain.ProviderConfig)}
}

func (s *MemoryStore) GetProviderConfig(_ context.Context, deviceID string) (domain.ProviderConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cfg, ok := s.providers[deviceID]
	if !ok {
		return domain.ProviderConfig{}, ErrNotFound
	}
	return cfg, nil
}

func (s *MemoryStore) SaveProviderConfig(_ context.Context, cfg domain.ProviderConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	cfg.UpdatedAt = time.Now().UTC()
	s.providers[cfg.DeviceID] = cfg
	return nil
}

func (s *MemoryStore) DeleteProviderConfig(_ context.Context, deviceID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.providers, deviceID)
	return nil
}

func (s *MemoryStore) SaveRequestLog(_ context.Context, log domain.RequestLog) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logs = append(s.logs, log)
	return nil
}

func (s *MemoryStore) ListRequestLogs(_ context.Context, deviceID string, limit int) ([]domain.RequestLog, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	out := make([]domain.RequestLog, 0, limit)
	for i := len(s.logs) - 1; i >= 0 && len(out) < limit; i-- {
		if deviceID == "" || s.logs[i].DeviceID == deviceID {
			out = append(out, s.logs[i])
		}
	}
	return out, nil
}

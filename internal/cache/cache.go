package cache

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"app/internal/metrics"
	"app/internal/models"
	"app/internal/repository"
)

type Decorator struct {
	repo  repository.UserProvider
	ttl   time.Duration
	mu    sync.RWMutex
	users map[string]*cacheEntry
}

type cacheEntry struct {
	user      *models.User
	expiredAt time.Time
}

func NewDecorator(repo repository.UserProvider, ttl time.Duration) *Decorator {
	return &Decorator{
		repo:  repo,
		ttl:   ttl,
		users: make(map[string]*cacheEntry),
	}
}

func (c *Decorator) set(user *models.User) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.users[user.ID] = &cacheEntry{
		user:      user,
		expiredAt: time.Now().Add(c.ttl),
	}
}

func (c *Decorator) get(id string) (*models.User, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.users[id]
	if !ok || time.Now().After(entry.expiredAt) {
		return nil, false
	}
	return entry.user, true
}

func (c *Decorator) Get(ctx context.Context, id string) (*models.User, error) {
	if user, ok := c.get(id); ok {
		metrics.IncCacheHits()
		slog.Debug("Cache hit", "userID", id)
		return user, nil
	}
	metrics.IncCacheMisses()
	userFromRepo, err := c.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	c.set(userFromRepo)
	slog.Debug("Cache miss - loaded from repo", "userID", id)
	return userFromRepo, nil
}

func (c *Decorator) CleanupExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for id, entry := range c.users {
		if now.After(entry.expiredAt) {
			delete(c.users, id)
			metrics.IncCacheEvictions()
			slog.Debug("Cache expired - user removed", "userID", id)
		}
	}
}

func (c *Decorator) GetAll(ctx context.Context, limit, offset int) ([]*models.User, error) {
	return c.repo.GetAll(ctx, limit, offset)
}

func (c *Decorator) Create(ctx context.Context, user *models.User) (string, error) {
	id, err := c.repo.Create(ctx, user)
	if err != nil {
		return "", err
	}
	user.ID = id
	c.set(user)
	return id, nil
}

func (c *Decorator) Update(ctx context.Context, user *models.User) error {
	if err := c.repo.Update(ctx, user); err != nil {
		return err
	}
	c.set(user)
	return nil
}

func (c *Decorator) delete(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.users, id)
}

func (c *Decorator) Delete(ctx context.Context, id string) error {
	if err := c.repo.Delete(ctx, id); err != nil {
		return err
	}
	c.delete(id)
	return nil
}

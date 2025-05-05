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
	repo         repository.UserProvider
	ttl          time.Duration
	mu           sync.RWMutex
	users        map[string]*cacheEntry
	cacheMetrics *metrics.CacheMetrics
}

type cacheEntry struct {
	user      *models.User
	expiredAt time.Time
}

func NewDecorator(repo repository.UserProvider, ttl time.Duration, metrics *metrics.CacheMetrics) *Decorator {
	return &Decorator{
		repo:         repo,
		ttl:          ttl,
		users:        make(map[string]*cacheEntry),
		cacheMetrics: metrics,
	}
}

func (c *Decorator) setToCache(user *models.User) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.users[user.ID] = &cacheEntry{
		user:      user,
		expiredAt: time.Now().Add(c.ttl),
	}
}

func (c *Decorator) getFromCache(id string) (*models.User, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.users[id]
	if !ok || time.Now().After(entry.expiredAt) {
		return entry.user, true
	}
	return nil, false
}

func (c *Decorator) Get(id string) (*models.User, error) {
	if user, ok := c.getFromCache(id); ok {
		c.cacheMetrics.Hits.Inc()
		slog.Debug("Cache hit", "userID", id)
		return user, nil
	}

	c.cacheMetrics.Misses.Inc()
	userFromRepo, err := c.repo.Get(id)
	if err != nil {
		return nil, err
	}

	c.setToCache(userFromRepo)
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
			c.cacheMetrics.Evictions.Inc()
			slog.Info("Cache expired - user removed", "userID", id)
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
	c.setToCache(user)
	return id, nil
}

func (c *Decorator) Update(ctx context.Context, user *models.User) error {
	if err := c.repo.Update(ctx, user); err != nil {
		return err
	}
	c.setToCache(user)
	return nil
}

func (c *Decorator) deleteFromCache(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.users, id)
}

func (c *Decorator) Delete(ctx context.Context, id string) error {
	if err := c.repo.Delete(ctx, id); err != nil {
		return err
	}
	c.deleteFromCache(id)
	return nil
}

package cache

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"app/internal/models"
	"app/internal/repository"

	"github.com/prometheus/client_golang/prometheus"
)

type Decorator struct {
	repo               repository.UserProvider
	mu                 sync.RWMutex
	data               map[string]*cacheEntry
	expirationDuration time.Duration
	cleanupInterval    time.Duration
	cacheHits          prometheus.Counter
	cacheMisses        prometheus.Counter
}

type cacheEntry struct {
	user      *models.User
	expiredAt time.Time
}

func NewDecorator(repo repository.UserProvider, expirationDuration, cleanupInterval time.Duration) *Decorator {
	cacheHits := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "cache_hits_total",
			Help: "Total number of cache hits",
		},
	)
	cacheMisses := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "cache_misses_total",
			Help: "Total number of cache misses",
		},
	)

	prometheus.MustRegister(cacheHits)
	prometheus.MustRegister(cacheMisses)

	return &Decorator{
		repo:               repo,
		data:               make(map[string]*cacheEntry),
		expirationDuration: expirationDuration,
		cleanupInterval:    cleanupInterval,
		cacheHits:          cacheHits,
		cacheMisses:        cacheMisses,
	}
}

func (c *Decorator) setToCache(user *models.User) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[user.ID] = &cacheEntry{
		user:      user,
		expiredAt: time.Now().Add(c.expirationDuration),
	}
}

func (c *Decorator) getFromCache(id string) (*models.User, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.data[id]
	if !ok || time.Now().After(entry.expiredAt) {
		return nil, false
	}
	return entry.user, true
}

func (c *Decorator) Get(id string) (*models.User, error) {
	if user, ok := c.getFromCache(id); ok {
		c.cacheHits.Inc()
		slog.Info("Cache hit", "userID", id)
		return user, nil
	}

	c.cacheMisses.Inc()
	userFromRepo, err := c.repo.Get(id)
	if err != nil {
		return nil, err
	}

	c.setToCache(userFromRepo)
	slog.Info("Cache miss - loaded from repo", "userID", id)
	return userFromRepo, nil
}

func (c *Decorator) CleanupExpiredLoop(ctx context.Context) {
	ticker := time.NewTicker(c.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			slog.Info("Cleaning expired cache entries...")
			c.cleanupExpiredCache()
		case <-ctx.Done():
			slog.Info("Cache cleanup loop shutting down")
			return
		}
	}
}

func (c *Decorator) cleanupExpiredCache() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for id, entry := range c.data {
		if now.After(entry.expiredAt) {
			slog.Info("Cache expired - user removed", "userID", id, "expiredAt", entry.expiredAt)
			delete(c.data, id)
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
	delete(c.data, id)
}

func (c *Decorator) Delete(ctx context.Context, id string) error {
	if err := c.repo.Delete(ctx, id); err != nil {
		return err
	}
	c.deleteFromCache(id)
	return nil
}

package cache

import (
	"context"
	"sync"
	"time"

	"app/internal/logger"
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

func (c *Decorator) Set(user *models.User) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[user.ID] = &cacheEntry{
		user:      user,
		expiredAt: time.Now().Add(c.expirationDuration),
	}
}

func (c *Decorator) Get(id string) (*models.User, error) {
	c.mu.RLock()
	entry, ok := c.data[id]
	c.mu.RUnlock()
	if ok && time.Now().Before(entry.expiredAt) {
		c.cacheHits.Inc()
		logger.Logger.Info("Cache hit", "userID", id)
		return entry.user, nil
	}

	c.cacheMisses.Inc()

	userFromRepo, err := c.repo.Get(id)
	if err != nil {
		return nil, err
	}

	c.Set(userFromRepo)
	logger.Logger.Info("Cache miss - loaded from repo", "userID", id)
	return userFromRepo, nil
}

func (c *Decorator) CleanupExpiredLoop(ctx context.Context) {
	ticker := time.NewTicker(c.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			logger.Logger.Info("Cleaning expired cache entries...")
			c.cleanupExpiredCache()
		case <-ctx.Done():
			logger.Logger.Info("Cache cleanup loop shutting down")
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
			logger.Logger.Info("Cache expired - user removed", "userID", id, "expiredAt", entry.expiredAt)
			delete(c.data, id)
		}
	}
}

func (c *Decorator) GetAll(limit, offset int) ([]*models.User, error) {
	return c.repo.GetAll(limit, offset)
}

func (c *Decorator) Create(user *models.User) (string, error) {
	id, err := c.repo.Create(user)
	if err != nil {
		return "", err
	}

	user.ID = id
	c.Set(user)
	return id, nil
}

func (c *Decorator) Update(user *models.User) error {
	err := c.repo.Update(user)
	if err != nil {
		return err
	}

	c.Set(user)
	return nil
}

func (c *Decorator) Delete(ctx context.Context, id string) error {
	err := c.repo.Delete(ctx, id)
	if err != nil {
		return err
	}

	c.mu.Lock()
	delete(c.data, id)
	c.mu.Unlock()
	return nil
}

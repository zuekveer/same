package cache

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"

	"app/internal/metrics"
	"app/internal/models"
	"app/internal/repository"
	"app/internal/tracing"
)

type Decorator struct {
	repo     repository.UserProvider
	ttl      time.Duration
	mu       sync.RWMutex
	users    map[string]*cacheEntry
	group    singleflight.Group
	groupAll singleflight.Group
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
	ctx, span := tracing.Start(ctx, "Cache.GetUser")
	defer span.End()

	if user, ok := c.get(id); ok {
		metrics.IncCacheHits()
		slog.Debug("Cache hit", "userID", id)
		return user, nil
	}

	metrics.IncCacheMisses()
	slog.Debug("Cache miss - loading from repo", "userID", id)

	result, err, _ := c.group.Do(id, func() (interface{}, error) {
		userFromRepo, err := c.repo.Get(ctx, id)
		if err != nil {
			return nil, err
		}
		c.set(userFromRepo)
		slog.Debug("Loaded from repo (singleflight)", "userID", id)
		return userFromRepo, nil
	})
	if err != nil {
		return nil, err
	}

	return result.(*models.User), nil
}

func (c *Decorator) CleanupExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for id, entry := range c.users {
		if now.After(entry.expiredAt) {
			delete(c.users, id)
			metrics.IncCacheExpired()
			slog.Debug("Cache expired - user removed", "userID", id)
		}
	}
}

func (c *Decorator) GetAll(ctx context.Context, limit, offset int) ([]*models.User, error) {
	ctx, span := tracing.Start(ctx, "Cache.GetAllUsers")
	defer span.End()

	key := fmt.Sprintf("getAll:%d:%d", limit, offset)
	result, err, _ := c.groupAll.Do(key, func() (interface{}, error) {
		return c.repo.GetAll(ctx, limit, offset)
	})
	if err != nil {
		return nil, err
	}
	return result.([]*models.User), nil
}

func (c *Decorator) Create(ctx context.Context, user *models.User) (string, error) {
	ctx, span := tracing.Start(ctx, "Cache.CreateUser")
	defer span.End()

	id, err := c.repo.Create(ctx, user)
	if err != nil {
		return "", err
	}
	user.ID = id
	c.set(user)
	return id, nil
}

func (c *Decorator) Update(ctx context.Context, user *models.User) error {
	ctx, span := tracing.Start(ctx, "Cache.UpdateUser")
	defer span.End()

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
	ctx, span := tracing.Start(ctx, "Cache.DeleteUser")
	defer span.End()

	if err := c.repo.Delete(ctx, id); err != nil {
		return err
	}
	c.delete(id)
	return nil
}

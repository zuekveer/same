package cache

import (
	"context"
	"sync"

	"app/internal/logger"
	"app/internal/models"
	"app/internal/repository"
)

type Decorator struct {
	repo repository.UserProvider
	mu   sync.RWMutex
	data map[string]models.User
}

func NewDecorator(repo repository.UserProvider) *Decorator {
	return &Decorator{
		repo: repo,
		data: make(map[string]models.User),
	}
}

func (c *Decorator) Get(id string) (*models.User, error) {
	c.mu.RLock()
	user, ok := c.data[id]
	c.mu.RUnlock()
	if ok {
		logger.Logger.Info("Cache hit", "userID", id)
		return &user, nil
	}

	user, err := c.repo.Get(id)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.data[id] = user
	c.mu.Unlock()
	logger.Logger.Info("Cache miss - loaded from repo", "userID", id)
	return &user, nil
}

func (c *Decorator) GetAll(limit, offset int) ([]*models.User, error) {
	return c.repo.GetAll(limit, offset)
}

func (c *Decorator) Create(user models.User) (string, error) {
	id, err := c.repo.Create(&user)
	if err == nil {
		c.mu.Lock()
		user.ID = id
		c.data[id] = user
		c.mu.Unlock()
	}
	return id, err
}

func (c *Decorator) Update(user *models.User) error {
	err := c.repo.Update(user)
	if err == nil {
		c.mu.Lock()
		c.data[user.ID] = *user
		c.mu.Unlock()
	}
	return err
}

func (c *Decorator) Delete(ctx context.Context, id string) error {
	err := c.repo.Delete(ctx, id)
	if err == nil {
		c.mu.Lock()
		delete(c.data, id)
		c.mu.Unlock()
	}
	return err
}

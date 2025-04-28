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
	data map[string]*models.User
}

func NewDecorator(repo repository.UserProvider) *Decorator {
	return &Decorator{
		repo: repo,
		data: make(map[string]*models.User),
	}
}

func (c *Decorator) Set(user *models.User) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[user.ID] = user
}

func (c *Decorator) Get(id string) (*models.User, error) {
	c.mu.RLock()
	user, ok := c.data[id]
	c.mu.RUnlock()
	if ok {
		logger.Logger.Info("Cache hit", "userID", id)
		return user, nil
	}

	userFromRepo, err := c.repo.Get(id)
	if err != nil {
		return nil, err
	}

	c.Set(userFromRepo)
	logger.Logger.Info("Cache miss - loaded from repo", "userID", id)
	return userFromRepo, nil
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

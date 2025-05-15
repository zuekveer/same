package cache

import (
	"context"
	"errors"
	"testing"
	"time"

	"app/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockUserProvider struct {
	mock.Mock
}

func (m *MockUserProvider) Create(ctx context.Context, user *models.User) (string, error) {
	args := m.Called(ctx, user)
	return args.String(0), args.Error(1)
}

func (m *MockUserProvider) Update(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserProvider) Get(ctx context.Context, id string) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserProvider) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserProvider) GetAll(ctx context.Context, limit, offset int) ([]*models.User, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func TestDecorator_Get_CacheHit(t *testing.T) {
	// Setup
	mockRepo := new(MockUserProvider)
	cache := NewDecorator(mockRepo, 10*time.Minute)

	testUser := &models.User{
		ID:   "123",
		Name: "Test User",
		Age:  30,
	}

	cache.set(testUser)

	user, err := cache.Get(context.Background(), "123")

	require.NoError(t, err)
	assert.Equal(t, testUser, user)
	mockRepo.AssertNotCalled(t, "Get")
}

func TestDecorator_Get_CacheMiss(t *testing.T) {
	mockRepo := new(MockUserProvider)
	cache := NewDecorator(mockRepo, 10*time.Minute)

	testUser := &models.User{
		ID:   "123",
		Name: "Test User",
		Age:  30,
	}

	mockRepo.On("Get", mock.Anything, "123").Return(testUser, nil).Once()

	user, err := cache.Get(context.Background(), "123")

	require.NoError(t, err)
	assert.Equal(t, testUser, user)
	mockRepo.AssertExpectations(t)

	cachedUser, ok := cache.get("123")
	assert.True(t, ok)
	assert.Equal(t, testUser, cachedUser)
}

func TestDecorator_Get_RepositoryError(t *testing.T) {
	mockRepo := new(MockUserProvider)
	cache := NewDecorator(mockRepo, 10*time.Minute)

	expectedErr := errors.New("repository error")

	mockRepo.On("Get", mock.Anything, "123").Return(nil, expectedErr).Once()

	user, err := cache.Get(context.Background(), "123")

	require.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, expectedErr, err)
	mockRepo.AssertExpectations(t)
}

func TestDecorator_Get_ExpiredEntry(t *testing.T) {
	mockRepo := new(MockUserProvider)
	cache := NewDecorator(mockRepo, 1*time.Nanosecond)

	testUser := &models.User{
		ID:   "123",
		Name: "Test User",
		Age:  30,
	}

	cache.set(testUser)
	time.Sleep(2 * time.Nanosecond) // Ensure entry expires

	mockRepo.On("Get", mock.Anything, "123").Return(testUser, nil).Once()

	user, err := cache.Get(context.Background(), "123")

	require.NoError(t, err)
	assert.Equal(t, testUser, user)
	mockRepo.AssertExpectations(t)
}

func TestDecorator_Create(t *testing.T) {
	mockRepo := new(MockUserProvider)
	cache := NewDecorator(mockRepo, 10*time.Minute)

	testUser := &models.User{
		Name: "Test User",
		Age:  30,
	}

	mockRepo.On("Create", mock.Anything, testUser).Return("123", nil).Once()

	id, err := cache.Create(context.Background(), testUser)

	require.NoError(t, err)
	assert.Equal(t, "123", id)
	mockRepo.AssertExpectations(t)

	cachedUser, ok := cache.get("123")
	assert.True(t, ok)
	assert.Equal(t, "123", cachedUser.ID)
	assert.Equal(t, testUser.Name, cachedUser.Name)
	assert.Equal(t, testUser.Age, cachedUser.Age)
}

func TestDecorator_Update(t *testing.T) {
	mockRepo := new(MockUserProvider)
	cache := NewDecorator(mockRepo, 10*time.Minute)

	testUser := &models.User{
		ID:   "123",
		Name: "Updated User",
		Age:  35,
	}

	mockRepo.On("Update", mock.Anything, testUser).Return(nil).Once()

	err := cache.Update(context.Background(), testUser)

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)

	cachedUser, ok := cache.get("123")
	assert.True(t, ok)
	assert.Equal(t, testUser, cachedUser)
}

func TestDecorator_Delete(t *testing.T) {
	mockRepo := new(MockUserProvider)
	cache := NewDecorator(mockRepo, 10*time.Minute)

	testUser := &models.User{
		ID:   "123",
		Name: "Test User",
		Age:  30,
	}

	cache.set(testUser)

	mockRepo.On("Delete", mock.Anything, "123").Return(nil).Once()

	err := cache.Delete(context.Background(), "123")

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)

	_, ok := cache.get("123")
	assert.False(t, ok)
}

func TestDecorator_GetAll(t *testing.T) {
	mockRepo := new(MockUserProvider)
	cache := NewDecorator(mockRepo, 10*time.Minute)

	testUsers := []*models.User{
		{ID: "1", Name: "User 1", Age: 30},
		{ID: "2", Name: "User 2", Age: 35},
	}

	mockRepo.On("GetAll", mock.Anything, 10, 0).Return(testUsers, nil).Once()

	users, err := cache.GetAll(context.Background(), 10, 0)

	require.NoError(t, err)
	assert.Equal(t, testUsers, users)
	mockRepo.AssertExpectations(t)

	_, ok := cache.get("1")
	assert.False(t, ok)
}

func TestDecorator_CleanupExpired(t *testing.T) {
	mockRepo := new(MockUserProvider)
	cache := NewDecorator(mockRepo, 1*time.Nanosecond)

	testUser := &models.User{
		ID:   "123",
		Name: "Test User",
		Age:  30,
	}

	cache.set(testUser)
	time.Sleep(2 * time.Nanosecond)

	cache.CleanupExpired()

	_, ok := cache.get("123")
	assert.False(t, ok)
}

func TestDecorator_ConcurrentAccess(t *testing.T) {
	mockRepo := new(MockUserProvider)
	cache := NewDecorator(mockRepo, 10*time.Minute)

	testUser := &models.User{
		ID:   "123",
		Name: "Test User",
		Age:  30,
	}

	mockRepo.On("Get", mock.Anything, "123").Return(testUser, nil).Run(func(args mock.Arguments) {
		time.Sleep(time.Millisecond * 10)
	}).Once()

	numGoroutines := 100

	results := make(chan *models.User, numGoroutines)
	errs := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			user, err := cache.Get(context.Background(), "123")
			if err != nil {
				errs <- err
				return
			}
			results <- user
		}()
	}

	for i := 0; i < numGoroutines; i++ {
		select {
		case user := <-results:
			assert.Equal(t, testUser, user)
		case err := <-errs:
			require.NoError(t, err)
		}
	}

	mockRepo.AssertExpectations(t)
}

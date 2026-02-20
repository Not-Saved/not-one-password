package mocks

import (
	"context"
	"fmt"
	"main/internal/core/domain"
	"sync"
	"time"
)

// MockUserRepository is an in-memory mock implementation of ports.UserRepository.
type MockUserRepository struct {
	mu     sync.Mutex
	users  []domain.User
	nextID int32

	// Error overrides — when set, the corresponding method returns this error.
	ListUsersErr  error
	GetUserErr    error
	CreateUserErr error

	// Call counters for verification.
	ListUsersCalled  int
	GetUserCalled    int
	CreateUserCalled int
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users:  make([]domain.User, 0),
		nextID: 1,
	}
}

func (m *MockUserRepository) ListUsers(ctx context.Context) ([]domain.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ListUsersCalled++

	if m.ListUsersErr != nil {
		return nil, m.ListUsersErr
	}

	result := make([]domain.User, len(m.users))
	copy(result, m.users)
	return result, nil
}

func (m *MockUserRepository) GetUser(ctx context.Context, id int32) (domain.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.GetUserCalled++

	if m.GetUserErr != nil {
		return domain.User{}, m.GetUserErr
	}

	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return domain.User{}, fmt.Errorf("user with id %d not found", id)
}

func (m *MockUserRepository) CreateUser(ctx context.Context, name, email string) (domain.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.CreateUserCalled++

	if m.CreateUserErr != nil {
		return domain.User{}, m.CreateUserErr
	}

	user := domain.User{
		ID:        m.nextID,
		Name:      name,
		Email:     email,
		CreatedAt: time.Now(),
	}
	m.nextID++
	m.users = append(m.users, user)
	return user, nil
}

// SeedUser adds a user directly into the mock store for test setup.
func (m *MockUserRepository) SeedUser(user domain.User) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.users = append(m.users, user)
	if user.ID >= m.nextID {
		m.nextID = user.ID + 1
	}
}

// Reset clears all state in the mock.
func (m *MockUserRepository) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.users = make([]domain.User, 0)
	m.nextID = 1
	m.ListUsersErr = nil
	m.GetUserErr = nil
	m.CreateUserErr = nil
	m.ListUsersCalled = 0
	m.GetUserCalled = 0
	m.CreateUserCalled = 0
}

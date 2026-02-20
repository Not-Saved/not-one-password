package services

import (
	"context"
	"fmt"
	"main/internal/core/domain"
	"testing"
	"time"
)

// mockUserRepository is a test double implementing ports.UserRepository.
type mockUserRepository struct {
	users  []domain.User
	nextID int32

	listUsersErr  error
	getUserErr    error
	createUserErr error
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users:  make([]domain.User, 0),
		nextID: 1,
	}
}

func (m *mockUserRepository) ListUsers(ctx context.Context) ([]domain.User, error) {
	if m.listUsersErr != nil {
		return nil, m.listUsersErr
	}
	result := make([]domain.User, len(m.users))
	copy(result, m.users)
	return result, nil
}

func (m *mockUserRepository) GetUser(ctx context.Context, id int32) (domain.User, error) {
	if m.getUserErr != nil {
		return domain.User{}, m.getUserErr
	}
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return domain.User{}, fmt.Errorf("user with id %d not found", id)
}

func (m *mockUserRepository) CreateUser(ctx context.Context, name, email string) (domain.User, error) {
	if m.createUserErr != nil {
		return domain.User{}, m.createUserErr
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

func (m *mockUserRepository) seedUser(user domain.User) {
	m.users = append(m.users, user)
	if user.ID >= m.nextID {
		m.nextID = user.ID + 1
	}
}

// ---------- Tests ----------

func TestNewUserService(t *testing.T) {
	repo := newMockUserRepository()
	svc := NewUserService(repo)

	if svc == nil {
		t.Fatal("expected non-nil UserService")
	}
}

func TestListUsers_Empty(t *testing.T) {
	repo := newMockUserRepository()
	svc := NewUserService(repo)

	users, err := svc.ListUsers(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(users) != 0 {
		t.Fatalf("expected 0 users, got %d", len(users))
	}
}

func TestListUsers_MultipleUsers(t *testing.T) {
	repo := newMockUserRepository()
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	repo.seedUser(domain.User{ID: 1, Name: "Alice", Email: "alice@example.com", CreatedAt: now})
	repo.seedUser(domain.User{ID: 2, Name: "Bob", Email: "bob@example.com", CreatedAt: now})
	repo.seedUser(domain.User{ID: 3, Name: "Charlie", Email: "charlie@example.com", CreatedAt: now})

	svc := NewUserService(repo)

	users, err := svc.ListUsers(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(users) != 3 {
		t.Fatalf("expected 3 users, got %d", len(users))
	}
	if users[0].Name != "Alice" {
		t.Errorf("expected first user name 'Alice', got %q", users[0].Name)
	}
	if users[1].Name != "Bob" {
		t.Errorf("expected second user name 'Bob', got %q", users[1].Name)
	}
	if users[2].Name != "Charlie" {
		t.Errorf("expected third user name 'Charlie', got %q", users[2].Name)
	}
}

func TestListUsers_RepoError(t *testing.T) {
	repo := newMockUserRepository()
	repo.listUsersErr = fmt.Errorf("database connection lost")

	svc := NewUserService(repo)

	users, err := svc.ListUsers(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "database connection lost" {
		t.Errorf("expected error message 'database connection lost', got %q", err.Error())
	}
	if users != nil {
		t.Errorf("expected nil users on error, got %v", users)
	}
}

func TestGetUser_Found(t *testing.T) {
	repo := newMockUserRepository()
	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	repo.seedUser(domain.User{ID: 42, Name: "Jane", Email: "jane@example.com", CreatedAt: now})

	svc := NewUserService(repo)

	user, err := svc.GetUser(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID != 42 {
		t.Errorf("expected ID 42, got %d", user.ID)
	}
	if user.Name != "Jane" {
		t.Errorf("expected name 'Jane', got %q", user.Name)
	}
	if user.Email != "jane@example.com" {
		t.Errorf("expected email 'jane@example.com', got %q", user.Email)
	}
	if !user.CreatedAt.Equal(now) {
		t.Errorf("expected CreatedAt %v, got %v", now, user.CreatedAt)
	}
}

func TestGetUser_NotFound(t *testing.T) {
	repo := newMockUserRepository()
	svc := NewUserService(repo)

	_, err := svc.GetUser(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for non-existent user, got nil")
	}
}

func TestGetUser_RepoError(t *testing.T) {
	repo := newMockUserRepository()
	repo.getUserErr = fmt.Errorf("timeout")

	svc := NewUserService(repo)

	_, err := svc.GetUser(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "timeout" {
		t.Errorf("expected error 'timeout', got %q", err.Error())
	}
}

func TestCreateUser_Success(t *testing.T) {
	repo := newMockUserRepository()
	svc := NewUserService(repo)

	user, err := svc.CreateUser(context.Background(), "NewUser", "new@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID == 0 {
		t.Error("expected non-zero ID")
	}
	if user.Name != "NewUser" {
		t.Errorf("expected name 'NewUser', got %q", user.Name)
	}
	if user.Email != "new@example.com" {
		t.Errorf("expected email 'new@example.com', got %q", user.Email)
	}
	if user.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
}

func TestCreateUser_MultiplePersists(t *testing.T) {
	repo := newMockUserRepository()
	svc := NewUserService(repo)

	u1, err := svc.CreateUser(context.Background(), "First", "first@example.com")
	if err != nil {
		t.Fatalf("unexpected error creating first user: %v", err)
	}

	u2, err := svc.CreateUser(context.Background(), "Second", "second@example.com")
	if err != nil {
		t.Fatalf("unexpected error creating second user: %v", err)
	}

	if u1.ID == u2.ID {
		t.Errorf("expected different IDs, both got %d", u1.ID)
	}

	// Verify both appear in list
	users, err := svc.ListUsers(context.Background())
	if err != nil {
		t.Fatalf("unexpected error listing users: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}
}

func TestCreateUser_RepoError(t *testing.T) {
	repo := newMockUserRepository()
	repo.createUserErr = fmt.Errorf("duplicate email")

	svc := NewUserService(repo)

	_, err := svc.CreateUser(context.Background(), "Dup", "dup@example.com")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "duplicate email" {
		t.Errorf("expected error 'duplicate email', got %q", err.Error())
	}
}

func TestCreateUser_EmptyFields(t *testing.T) {
	repo := newMockUserRepository()
	svc := NewUserService(repo)

	// The service currently does no validation — it delegates to the repo.
	// This test documents that behaviour.
	user, err := svc.CreateUser(context.Background(), "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Name != "" {
		t.Errorf("expected empty name, got %q", user.Name)
	}
	if user.Email != "" {
		t.Errorf("expected empty email, got %q", user.Email)
	}
}

func TestCreateThenGet(t *testing.T) {
	repo := newMockUserRepository()
	svc := NewUserService(repo)

	created, err := svc.CreateUser(context.Background(), "Roundtrip", "rt@example.com")
	if err != nil {
		t.Fatalf("unexpected error creating user: %v", err)
	}

	fetched, err := svc.GetUser(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("unexpected error getting user: %v", err)
	}

	if created.ID != fetched.ID {
		t.Errorf("ID mismatch: created %d, fetched %d", created.ID, fetched.ID)
	}
	if created.Name != fetched.Name {
		t.Errorf("Name mismatch: created %q, fetched %q", created.Name, fetched.Name)
	}
	if created.Email != fetched.Email {
		t.Errorf("Email mismatch: created %q, fetched %q", created.Email, fetched.Email)
	}
}

func TestContextCancellation(t *testing.T) {
	repo := newMockUserRepository()
	svc := NewUserService(repo)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	// The mock doesn't check context, so these should still succeed.
	// This test documents the current behaviour and ensures no panics.
	_, err := svc.ListUsers(ctx)
	if err != nil {
		t.Fatalf("unexpected error with cancelled context: %v", err)
	}

	_, err = svc.CreateUser(ctx, "CancelTest", "cancel@example.com")
	if err != nil {
		t.Fatalf("unexpected error with cancelled context: %v", err)
	}
}

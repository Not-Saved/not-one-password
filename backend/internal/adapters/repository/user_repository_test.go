package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"main/internal/core/domain"
)

func TestNewUserRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewUserRepository(db)
	if repo == nil {
		t.Fatal("expected non-nil UserRepository")
	}
}

// ---------- ListUsers ----------

func TestListUsers_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "name", "email", "created_at"})
	mock.ExpectQuery("SELECT id, name, email, created_at FROM users").
		WillReturnRows(rows)

	repo := NewUserRepository(db)
	users, err := repo.ListUsers(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(users) != 0 {
		t.Errorf("expected 0 users, got %d", len(users))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestListUsers_MultipleUsers(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	rows := sqlmock.NewRows([]string{"id", "name", "email", "created_at"}).
		AddRow(1, "Alice", "alice@example.com", now).
		AddRow(2, "Bob", "bob@example.com", now).
		AddRow(3, "Charlie", "charlie@example.com", now)

	mock.ExpectQuery("SELECT id, name, email, created_at FROM users").
		WillReturnRows(rows)

	repo := NewUserRepository(db)
	users, err := repo.ListUsers(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(users) != 3 {
		t.Fatalf("expected 3 users, got %d", len(users))
	}

	expected := []domain.User{
		{ID: 1, Name: "Alice", Email: "alice@example.com", CreatedAt: now},
		{ID: 2, Name: "Bob", Email: "bob@example.com", CreatedAt: now},
		{ID: 3, Name: "Charlie", Email: "charlie@example.com", CreatedAt: now},
	}

	for i, exp := range expected {
		if users[i].ID != exp.ID {
			t.Errorf("user %d: expected ID %d, got %d", i, exp.ID, users[i].ID)
		}
		if users[i].Name != exp.Name {
			t.Errorf("user %d: expected Name %q, got %q", i, exp.Name, users[i].Name)
		}
		if users[i].Email != exp.Email {
			t.Errorf("user %d: expected Email %q, got %q", i, exp.Email, users[i].Email)
		}
		if !users[i].CreatedAt.Equal(exp.CreatedAt) {
			t.Errorf("user %d: expected CreatedAt %v, got %v", i, exp.CreatedAt, users[i].CreatedAt)
		}
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestListUsers_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT id, name, email, created_at FROM users").
		WillReturnError(fmt.Errorf("connection refused"))

	repo := NewUserRepository(db)
	users, err := repo.ListUsers(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if users != nil {
		t.Errorf("expected nil users on error, got %v", users)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestListUsers_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	// Return a row with a wrong type for id (string instead of int32) to trigger scan error
	rows := sqlmock.NewRows([]string{"id", "name", "email", "created_at"}).
		AddRow("not_a_number", "Alice", "alice@example.com", time.Now())

	mock.ExpectQuery("SELECT id, name, email, created_at FROM users").
		WillReturnRows(rows)

	repo := NewUserRepository(db)
	_, err = repo.ListUsers(context.Background())
	if err == nil {
		t.Fatal("expected scan error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// ---------- GetUser ----------

func TestGetUser_Found(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	now := time.Date(2025, 3, 20, 10, 30, 0, 0, time.UTC)
	rows := sqlmock.NewRows([]string{"id", "name", "email", "created_at"}).
		AddRow(42, "Jane", "jane@example.com", now)

	mock.ExpectQuery("SELECT id, name, email, created_at FROM users").
		WithArgs(int32(42)).
		WillReturnRows(rows)

	repo := NewUserRepository(db)
	user, err := repo.GetUser(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if user.ID != 42 {
		t.Errorf("expected ID 42, got %d", user.ID)
	}
	if user.Name != "Jane" {
		t.Errorf("expected Name 'Jane', got %q", user.Name)
	}
	if user.Email != "jane@example.com" {
		t.Errorf("expected Email 'jane@example.com', got %q", user.Email)
	}
	if !user.CreatedAt.Equal(now) {
		t.Errorf("expected CreatedAt %v, got %v", now, user.CreatedAt)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestGetUser_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "name", "email", "created_at"})
	mock.ExpectQuery("SELECT id, name, email, created_at FROM users").
		WithArgs(int32(999)).
		WillReturnRows(rows)

	repo := NewUserRepository(db)
	user, err := repo.GetUser(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for non-existent user, got nil")
	}

	// Should return a zero-value user
	if user.ID != 0 {
		t.Errorf("expected zero-value user ID, got %d", user.ID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestGetUser_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT id, name, email, created_at FROM users").
		WithArgs(int32(1)).
		WillReturnError(fmt.Errorf("database timeout"))

	repo := NewUserRepository(db)
	_, err = repo.GetUser(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "database timeout" {
		t.Errorf("expected 'database timeout', got %q", err.Error())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// ---------- CreateUser ----------

func TestCreateUser_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	now := time.Date(2025, 7, 1, 8, 0, 0, 0, time.UTC)
	rows := sqlmock.NewRows([]string{"id", "name", "email", "created_at"}).
		AddRow(1, "NewUser", "new@example.com", now)

	mock.ExpectQuery("INSERT INTO users").
		WithArgs("NewUser", "new@example.com").
		WillReturnRows(rows)

	repo := NewUserRepository(db)
	user, err := repo.CreateUser(context.Background(), "NewUser", "new@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if user.ID != 1 {
		t.Errorf("expected ID 1, got %d", user.ID)
	}
	if user.Name != "NewUser" {
		t.Errorf("expected Name 'NewUser', got %q", user.Name)
	}
	if user.Email != "new@example.com" {
		t.Errorf("expected Email 'new@example.com', got %q", user.Email)
	}
	if !user.CreatedAt.Equal(now) {
		t.Errorf("expected CreatedAt %v, got %v", now, user.CreatedAt)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestCreateUser_DuplicateEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("INSERT INTO users").
		WithArgs("Dup", "dup@example.com").
		WillReturnError(fmt.Errorf("unique constraint violation: email"))

	repo := NewUserRepository(db)
	user, err := repo.CreateUser(context.Background(), "Dup", "dup@example.com")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if user.ID != 0 {
		t.Errorf("expected zero-value user on error, got ID %d", user.ID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestCreateUser_EmptyFields(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	now := time.Date(2025, 7, 1, 8, 0, 0, 0, time.UTC)
	rows := sqlmock.NewRows([]string{"id", "name", "email", "created_at"}).
		AddRow(1, "", "", now)

	mock.ExpectQuery("INSERT INTO users").
		WithArgs("", "").
		WillReturnRows(rows)

	repo := NewUserRepository(db)
	user, err := repo.CreateUser(context.Background(), "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if user.Name != "" {
		t.Errorf("expected empty name, got %q", user.Name)
	}
	if user.Email != "" {
		t.Errorf("expected empty email, got %q", user.Email)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestCreateUser_SpecialCharacters(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	now := time.Date(2025, 7, 1, 8, 0, 0, 0, time.UTC)
	name := "O'Brien & Co."
	email := "o'brien+tag@example.com"
	rows := sqlmock.NewRows([]string{"id", "name", "email", "created_at"}).
		AddRow(1, name, email, now)

	mock.ExpectQuery("INSERT INTO users").
		WithArgs(name, email).
		WillReturnRows(rows)

	repo := NewUserRepository(db)
	user, err := repo.CreateUser(context.Background(), name, email)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if user.Name != name {
		t.Errorf("expected Name %q, got %q", name, user.Name)
	}
	if user.Email != email {
		t.Errorf("expected Email %q, got %q", email, user.Email)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// ---------- toDomainUser mapping ----------

func TestToDomainUser(t *testing.T) {
	// We can't import db package types directly in a black-box test from
	// this package, but toDomainUser is in the same package so we test it
	// indirectly via the repository methods. The GetUser/CreateUser/ListUsers
	// tests above already cover the mapping. This test is here as explicit
	// documentation that the mapping preserves all fields.

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	createdAt := time.Date(2020, 2, 29, 23, 59, 59, 0, time.UTC)
	rows := sqlmock.NewRows([]string{"id", "name", "email", "created_at"}).
		AddRow(int32(2147483647), "MaxID User", "maxid@example.com", createdAt)

	mock.ExpectQuery("SELECT id, name, email, created_at FROM users").
		WithArgs(int32(2147483647)).
		WillReturnRows(rows)

	repo := NewUserRepository(db)
	user, err := repo.GetUser(context.Background(), 2147483647)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if user.ID != 2147483647 {
		t.Errorf("expected max int32 ID, got %d", user.ID)
	}
	if user.Name != "MaxID User" {
		t.Errorf("expected Name 'MaxID User', got %q", user.Name)
	}
	if user.Email != "maxid@example.com" {
		t.Errorf("expected Email 'maxid@example.com', got %q", user.Email)
	}
	if !user.CreatedAt.Equal(createdAt) {
		t.Errorf("expected CreatedAt %v, got %v", createdAt, user.CreatedAt)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// ---------- Context propagation ----------

func TestListUsers_CancelledContext(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// sqlmock doesn't enforce context cancellation, so the query will still
	// match. This test ensures the repository doesn't panic on a cancelled context.
	rows := sqlmock.NewRows([]string{"id", "name", "email", "created_at"})
	mock.ExpectQuery("SELECT id, name, email, created_at FROM users").
		WillReturnRows(rows)

	repo := NewUserRepository(db)
	_, err = repo.ListUsers(ctx)
	// Either success or context error is acceptable; we just ensure no panic.
	_ = err

	// We don't check expectations here because context cancellation may
	// prevent the query from being executed.
}

func TestCreateUser_ConnectionError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("INSERT INTO users").
		WithArgs("Test", "test@example.com").
		WillReturnError(fmt.Errorf("connection reset by peer"))

	repo := NewUserRepository(db)
	user, err := repo.CreateUser(context.Background(), "Test", "test@example.com")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "connection reset by peer" {
		t.Errorf("expected 'connection reset by peer', got %q", err.Error())
	}
	if user != (domain.User{}) {
		t.Errorf("expected zero-value user on error, got %+v", user)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

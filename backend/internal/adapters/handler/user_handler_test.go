package handler

import (
	"context"
	"fmt"
	"main/internal/core/domain"
	"main/internal/core/services"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

// mockUserRepository implements ports.UserRepository for testing handlers.
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
		CreatedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
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

func setupHandler(repo *mockUserRepository) *UserHandler {
	svc := services.NewUserService(repo)
	return NewUserHandler(svc)
}

// ---------- NewUserHandler ----------

func TestNewUserHandler(t *testing.T) {
	repo := newMockUserRepository()
	h := setupHandler(repo)
	if h == nil {
		t.Fatal("expected non-nil UserHandler")
	}
}

// ---------- ListUsers ----------

func TestListUsers_Empty(t *testing.T) {
	repo := newMockUserRepository()
	h := setupHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rec := httptest.NewRecorder()

	h.ListUsers(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if body != "" {
		t.Errorf("expected empty body for no users, got %q", body)
	}
}

func TestListUsers_SingleUser(t *testing.T) {
	repo := newMockUserRepository()
	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	repo.seedUser(domain.User{ID: 1, Name: "Alice", Email: "alice@example.com", CreatedAt: now})

	h := setupHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rec := httptest.NewRecorder()

	h.ListUsers(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	expected := "1 - Alice (alice@example.com)\n"
	if rec.Body.String() != expected {
		t.Errorf("expected body %q, got %q", expected, rec.Body.String())
	}
}

func TestListUsers_MultipleUsers(t *testing.T) {
	repo := newMockUserRepository()
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	repo.seedUser(domain.User{ID: 1, Name: "Alice", Email: "alice@example.com", CreatedAt: now})
	repo.seedUser(domain.User{ID: 2, Name: "Bob", Email: "bob@example.com", CreatedAt: now})
	repo.seedUser(domain.User{ID: 3, Name: "Charlie", Email: "charlie@example.com", CreatedAt: now})

	h := setupHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rec := httptest.NewRecorder()

	h.ListUsers(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	lines := strings.Split(strings.TrimSpace(body), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d: %q", len(lines), body)
	}

	expectedLines := []string{
		"1 - Alice (alice@example.com)",
		"2 - Bob (bob@example.com)",
		"3 - Charlie (charlie@example.com)",
	}
	for i, expected := range expectedLines {
		if lines[i] != expected {
			t.Errorf("line %d: expected %q, got %q", i, expected, lines[i])
		}
	}
}

func TestListUsers_RepoError(t *testing.T) {
	repo := newMockUserRepository()
	repo.listUsersErr = fmt.Errorf("database connection lost")

	h := setupHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rec := httptest.NewRecorder()

	h.ListUsers(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}

	body := strings.TrimSpace(rec.Body.String())
	if body != "database connection lost" {
		t.Errorf("expected error message 'database connection lost', got %q", body)
	}
}

// ---------- CreateUser ----------

func TestCreateUser_Success(t *testing.T) {
	repo := newMockUserRepository()
	h := setupHandler(repo)

	form := url.Values{}
	form.Set("name", "NewUser")
	form.Set("email", "new@example.com")

	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	h.CreateUser(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	expected := "Created user: 1 - NewUser (new@example.com)\n"
	if rec.Body.String() != expected {
		t.Errorf("expected body %q, got %q", expected, rec.Body.String())
	}
}

func TestCreateUser_RepoError(t *testing.T) {
	repo := newMockUserRepository()
	repo.createUserErr = fmt.Errorf("duplicate email constraint")

	h := setupHandler(repo)

	form := url.Values{}
	form.Set("name", "Dup")
	form.Set("email", "dup@example.com")

	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	h.CreateUser(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}

	body := strings.TrimSpace(rec.Body.String())
	if body != "duplicate email constraint" {
		t.Errorf("expected error message 'duplicate email constraint', got %q", body)
	}
}

func TestCreateUser_EmptyFields(t *testing.T) {
	repo := newMockUserRepository()
	h := setupHandler(repo)

	// Submit with no form values — name and email will be empty strings
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	h.CreateUser(rec, req)

	// The handler currently does no validation, so it should succeed
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	expected := "Created user: 1 -  ()\n"
	if rec.Body.String() != expected {
		t.Errorf("expected body %q, got %q", expected, rec.Body.String())
	}
}

func TestCreateUser_PartialFields(t *testing.T) {
	repo := newMockUserRepository()
	h := setupHandler(repo)

	form := url.Values{}
	form.Set("name", "OnlyName")
	// email intentionally omitted

	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	h.CreateUser(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	expected := "Created user: 1 - OnlyName ()\n"
	if rec.Body.String() != expected {
		t.Errorf("expected body %q, got %q", expected, rec.Body.String())
	}
}

func TestCreateUser_MultipleCreations(t *testing.T) {
	repo := newMockUserRepository()
	h := setupHandler(repo)

	// Create first user
	form1 := url.Values{}
	form1.Set("name", "First")
	form1.Set("email", "first@example.com")

	req1 := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(form1.Encode()))
	req1.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec1 := httptest.NewRecorder()
	h.CreateUser(rec1, req1)

	if rec1.Code != http.StatusOK {
		t.Fatalf("first create: expected status 200, got %d", rec1.Code)
	}

	// Create second user
	form2 := url.Values{}
	form2.Set("name", "Second")
	form2.Set("email", "second@example.com")

	req2 := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(form2.Encode()))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec2 := httptest.NewRecorder()
	h.CreateUser(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Fatalf("second create: expected status 200, got %d", rec2.Code)
	}

	// IDs should be different
	if rec1.Body.String() == rec2.Body.String() {
		t.Error("expected different responses for two different users")
	}

	expected2 := "Created user: 2 - Second (second@example.com)\n"
	if rec2.Body.String() != expected2 {
		t.Errorf("expected body %q, got %q", expected2, rec2.Body.String())
	}
}

func TestCreateUser_SpecialCharacters(t *testing.T) {
	repo := newMockUserRepository()
	h := setupHandler(repo)

	form := url.Values{}
	form.Set("name", "O'Brien & Co.")
	form.Set("email", "o'brien+tag@example.com")

	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	h.CreateUser(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "O'Brien & Co.") {
		t.Errorf("expected body to contain name with special characters, got %q", body)
	}
	if !strings.Contains(body, "o'brien+tag@example.com") {
		t.Errorf("expected body to contain email with special characters, got %q", body)
	}
}

// ---------- Integration-style: Create then List ----------

func TestCreateThenList(t *testing.T) {
	repo := newMockUserRepository()
	h := setupHandler(repo)

	// Create a user
	form := url.Values{}
	form.Set("name", "Integration")
	form.Set("email", "integration@example.com")

	createReq := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(form.Encode()))
	createReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	createRec := httptest.NewRecorder()
	h.CreateUser(createRec, createReq)

	if createRec.Code != http.StatusOK {
		t.Fatalf("create: expected status 200, got %d", createRec.Code)
	}

	// List users and verify the created user appears
	listReq := httptest.NewRequest(http.MethodGet, "/users", nil)
	listRec := httptest.NewRecorder()
	h.ListUsers(listRec, listReq)

	if listRec.Code != http.StatusOK {
		t.Fatalf("list: expected status 200, got %d", listRec.Code)
	}

	body := listRec.Body.String()
	if !strings.Contains(body, "Integration") {
		t.Errorf("expected listing to contain created user, got %q", body)
	}
	if !strings.Contains(body, "integration@example.com") {
		t.Errorf("expected listing to contain created user's email, got %q", body)
	}
}

// ---------- Content-Type / Response checks ----------

func TestListUsers_ResponseContentType(t *testing.T) {
	repo := newMockUserRepository()
	repo.seedUser(domain.User{
		ID: 1, Name: "Alice", Email: "alice@example.com",
		CreatedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	})

	h := setupHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rec := httptest.NewRecorder()

	h.ListUsers(rec, req)

	// The handler uses fmt.Fprintf, so default Content-Type detection applies.
	// httptest.ResponseRecorder defaults to 200 OK with text/plain when
	// Content-Type is detected by http.DetectContentType on first Write.
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestCreateUser_NoContentType(t *testing.T) {
	repo := newMockUserRepository()
	h := setupHandler(repo)

	// POST with no Content-Type header — FormValue should still return empty strings
	req := httptest.NewRequest(http.MethodPost, "/users", nil)
	rec := httptest.NewRecorder()

	h.CreateUser(rec, req)

	// Should succeed but with empty name/email
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

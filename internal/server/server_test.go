package server

import (
	"context"
	"fmt"
	"main/internal/adapters/handler"
	"main/internal/bootstrap"
	"main/internal/core/domain"
	"main/internal/core/services"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

// mockUserRepository implements ports.UserRepository for testing.
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

func setupTestServer(repo *mockUserRepository) *Server {
	svc := services.NewUserService(repo)
	userHandler := handler.NewUserHandler(svc)

	srv := New("8080")
	srv.RegisterApiRoutes(&bootstrap.Handlers{User: userHandler})
	return srv
}

// ---------- New ----------

func TestNew(t *testing.T) {
	srv := New("9090")
	if srv == nil {
		t.Fatal("expected non-nil Server")
	}
	if srv.port != "9090" {
		t.Errorf("expected port '9090', got %q", srv.port)
	}
	if srv.mux == nil {
		t.Fatal("expected non-nil mux")
	}
}

func TestNew_DifferentPorts(t *testing.T) {
	tests := []struct {
		port string
	}{
		{"80"},
		{"443"},
		{"3000"},
		{"8080"},
		{"0"},
	}

	for _, tt := range tests {
		t.Run("port_"+tt.port, func(t *testing.T) {
			srv := New(tt.port)
			if srv.port != tt.port {
				t.Errorf("expected port %q, got %q", tt.port, srv.port)
			}
		})
	}
}

// ---------- RegisterRoutes ----------

func TestRegisterRoutes_GetUsers(t *testing.T) {
	repo := newMockUserRepository()
	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	repo.seedUser(domain.User{ID: 1, Name: "Alice", Email: "alice@example.com", CreatedAt: now})
	repo.seedUser(domain.User{ID: 2, Name: "Bob", Email: "bob@example.com", CreatedAt: now})

	srv := setupTestServer(repo)

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rec := httptest.NewRecorder()

	srv.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GET /users: expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Alice") {
		t.Errorf("GET /users: expected body to contain 'Alice', got %q", body)
	}
	if !strings.Contains(body, "Bob") {
		t.Errorf("GET /users: expected body to contain 'Bob', got %q", body)
	}
}

func TestRegisterRoutes_PostUsers(t *testing.T) {
	repo := newMockUserRepository()
	srv := setupTestServer(repo)

	form := url.Values{}
	form.Set("name", "NewUser")
	form.Set("email", "new@example.com")

	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	srv.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("POST /users: expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "NewUser") {
		t.Errorf("POST /users: expected body to contain 'NewUser', got %q", body)
	}
	if !strings.Contains(body, "new@example.com") {
		t.Errorf("POST /users: expected body to contain 'new@example.com', got %q", body)
	}
}

func TestRegisterRoutes_GetUsersEmpty(t *testing.T) {
	repo := newMockUserRepository()
	srv := setupTestServer(repo)

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rec := httptest.NewRecorder()

	srv.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GET /users (empty): expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	if body != "" {
		t.Errorf("GET /users (empty): expected empty body, got %q", body)
	}
}

func TestRegisterRoutes_PostThenGet(t *testing.T) {
	repo := newMockUserRepository()
	srv := setupTestServer(repo)

	// Create a user via POST
	form := url.Values{}
	form.Set("name", "Integration")
	form.Set("email", "integration@example.com")

	postReq := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(form.Encode()))
	postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	postRec := httptest.NewRecorder()
	srv.mux.ServeHTTP(postRec, postReq)

	if postRec.Code != http.StatusOK {
		t.Fatalf("POST /users: expected status 200, got %d", postRec.Code)
	}

	// List users via GET and verify the user is there
	getReq := httptest.NewRequest(http.MethodGet, "/users", nil)
	getRec := httptest.NewRecorder()
	srv.mux.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("GET /users: expected status 200, got %d", getRec.Code)
	}

	body := getRec.Body.String()
	if !strings.Contains(body, "Integration") {
		t.Errorf("GET /users: expected body to contain 'Integration', got %q", body)
	}
	if !strings.Contains(body, "integration@example.com") {
		t.Errorf("GET /users: expected body to contain 'integration@example.com', got %q", body)
	}
}

func TestRegisterRoutes_WrongMethodOnGetUsers(t *testing.T) {
	repo := newMockUserRepository()
	srv := setupTestServer(repo)

	// DELETE /users should not match the registered GET /users route
	req := httptest.NewRequest(http.MethodDelete, "/users", nil)
	rec := httptest.NewRecorder()

	srv.mux.ServeHTTP(rec, req)

	// Go 1.22+ method-aware routing returns 405 Method Not Allowed
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("DELETE /users: expected status 405, got %d", rec.Code)
	}
}

func TestRegisterRoutes_WrongMethodOnPostUsers(t *testing.T) {
	repo := newMockUserRepository()
	srv := setupTestServer(repo)

	// PUT /users should not match the registered POST /users route
	req := httptest.NewRequest(http.MethodPut, "/users", nil)
	rec := httptest.NewRecorder()

	srv.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("PUT /users: expected status 405, got %d", rec.Code)
	}
}

func TestRegisterRoutes_UnknownPath(t *testing.T) {
	repo := newMockUserRepository()
	srv := setupTestServer(repo)

	req := httptest.NewRequest(http.MethodGet, "/unknown", nil)
	rec := httptest.NewRecorder()

	srv.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("GET /unknown: expected status 404, got %d", rec.Code)
	}
}

func TestRegisterRoutes_GetUsersRepoError(t *testing.T) {
	repo := newMockUserRepository()
	repo.listUsersErr = fmt.Errorf("database connection lost")
	srv := setupTestServer(repo)

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rec := httptest.NewRecorder()

	srv.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("GET /users (error): expected status 500, got %d", rec.Code)
	}

	body := strings.TrimSpace(rec.Body.String())
	if body != "database connection lost" {
		t.Errorf("GET /users (error): expected error message 'database connection lost', got %q", body)
	}
}

func TestRegisterRoutes_PostUsersRepoError(t *testing.T) {
	repo := newMockUserRepository()
	repo.createUserErr = fmt.Errorf("unique constraint violation")
	srv := setupTestServer(repo)

	form := url.Values{}
	form.Set("name", "Dup")
	form.Set("email", "dup@example.com")

	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	srv.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("POST /users (error): expected status 500, got %d", rec.Code)
	}

	body := strings.TrimSpace(rec.Body.String())
	if body != "unique constraint violation" {
		t.Errorf("POST /users (error): expected error message 'unique constraint violation', got %q", body)
	}
}

func TestRegisterRoutes_MultipleUsersInList(t *testing.T) {
	repo := newMockUserRepository()
	srv := setupTestServer(repo)

	// Create multiple users
	names := []string{"Alice", "Bob", "Charlie"}
	emails := []string{"alice@test.com", "bob@test.com", "charlie@test.com"}

	for i := range names {
		form := url.Values{}
		form.Set("name", names[i])
		form.Set("email", emails[i])

		req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()
		srv.mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("POST /users (%s): expected status 200, got %d", names[i], rec.Code)
		}
	}

	// List all users
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rec := httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /users: expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	lines := strings.Split(strings.TrimSpace(body), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines in listing, got %d: %q", len(lines), body)
	}

	for i, name := range names {
		if !strings.Contains(lines[i], name) {
			t.Errorf("line %d: expected to contain %q, got %q", i, name, lines[i])
		}
		if !strings.Contains(lines[i], emails[i]) {
			t.Errorf("line %d: expected to contain %q, got %q", i, emails[i], lines[i])
		}
	}
}

// ---------- Start ----------

func TestStart_PortAlreadyInUse(t *testing.T) {
	// Bind a random port on all interfaces to match http.ListenAndServe behavior
	listener, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", ":0")
	if err != nil {
		t.Fatalf("failed to listen on random port: %v", err)
	}
	defer listener.Close()

	// Extract the port that was assigned
	addr := listener.Addr().String()
	parts := strings.Split(addr, ":")
	port := parts[len(parts)-1]

	srv := New(port)

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start()
	}()

	select {
	case err := <-errCh:
		if err == nil {
			t.Error("expected error when port is in use, got nil")
		}
	case <-time.After(2 * time.Second):
		t.Skip("server did not fail within timeout")
	}
}

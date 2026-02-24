package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"main/internal/core/domain"
	"main/internal/core/services"
	"main/internal/oapi"

	"github.com/oapi-codegen/runtime/types"
)

type UserHandler struct {
	UserService *services.UserService
}

func NewUserHandler(service *services.UserService) *UserHandler {
	return &UserHandler{UserService: service}
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(oapi.ErrorResponse{
		Code:    status,
		Message: message,
	})
}

func mapToAPIUser(u domain.User) oapi.User {
	id := u.ID
	name := u.Name
	email := types.Email(u.Email)
	createdAt := u.CreatedAt

	return oapi.User{
		ID:        &id,
		Name:      &name,
		Email:     &email,
		CreatedAt: &createdAt,
	}
}

func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.UserService.ListUsers(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := make([]oapi.User, 0, len(users))

	for _, u := range users {
		response = append(response, mapToAPIUser(u))
	}

	json.NewEncoder(w).Encode(response)
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req oapi.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" || req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "missing required fields")
		return
	}

	user, err := h.UserService.CreateUser(r.Context(), req.Name, string(req.Email), req.Password)

	if err != nil {
		log.Printf("create user failed: %v", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := mapToAPIUser(user)

	json.NewEncoder(w).Encode(response)
}

func (h *UserHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	var req oapi.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	req.Email = types.Email(strings.TrimSpace(string(req.Email)))
	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "missing required fields")
		return
	}

	userAgent := r.UserAgent()
	ip := r.RemoteAddr

	user, session, err := h.UserService.LoginUser(r.Context(), string(req.Email), req.Password, userAgent, ip)

	if err != nil {
		log.Printf("login failed: %v", err)
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	response := mapToAPIUser(*user)

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    session.Token,
		HttpOnly: true,
		Secure:   false,
		Expires:  session.ExpiresAt,
	})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

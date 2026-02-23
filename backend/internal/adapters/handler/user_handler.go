package handler

import (
	"encoding/json"
	"net/http"

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
	if err := r.ParseForm(); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid form data")
		return
	}

	name := r.FormValue("name")
	email := r.FormValue("email")

	if name == "" || email == "" {
		writeError(w, http.StatusBadRequest, "Missing required fields: name, email")
		return
	}

	user, err := h.UserService.CreateUser(r.Context(), name, email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := mapToAPIUser(user)

	json.NewEncoder(w).Encode(response)
}

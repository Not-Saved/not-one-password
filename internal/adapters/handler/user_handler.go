package handler

import (
	"fmt"
	"main/internal/core/services"
	"net/http"
)

type UserHandler struct {
	service *services.UserService
}

func NewUserHandler(service *services.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.service.ListUsers(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, u := range users {
		fmt.Fprintf(w, "%d - %s (%s)\n", u.ID, u.Name, u.Email)
	}

}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	email := r.FormValue("email")

	user, err := h.service.CreateUser(r.Context(), name, email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Created user: %d - %s (%s)\n", user.ID, user.Name, user.Email)
}

package ports

import (
	"context"
	"main/internal/core/domain"
)

type UserIntentRepository interface {
	CreateRegistrationIntent(ctx context.Context, user domain.RegistrationIntentUser) (*domain.RegistrationIntentToken, error)
	GetRegistrationIntent(ctx context.Context, code string) (*domain.RegistrationIntentUser, error)
	DeleteRegistrationIntent(ctx context.Context, code string) error
}

package domain

type RegistrationIntentUser struct {
	Name         string
	PasswordHash string
	Email        string
	Code         string
}

type RegistrationIntentToken struct {
	Code string
}

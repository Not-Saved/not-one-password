package repository

import (
	"main/internal/utils"
	"time"
)

const (
	ACCESS_TOKEN_EXPIRATION  = 15 * time.Minute
	REFRESH_TOKEN_EXPIRATION = 7 * 24 * time.Hour
)

func GenerateTokenForSession() (string, string, error) {
	token, err := utils.GenerateRandomString(32)
	if err != nil {
		return "", "", err
	}

	return token, utils.HashToken(token), nil
}

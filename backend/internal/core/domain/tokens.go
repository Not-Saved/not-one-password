package domain

import (
	"encoding/base64"
	"encoding/json"
)

type Tokens struct {
	// AccessToken JWT access token
	AccessToken string `json:"accessToken"`

	// ExpiresIn Access token expiration time in seconds
	ExpiresIn int `json:"expiresIn"`

	// RefreshToken Refresh token
	RefreshToken string `json:"refreshToken"`
}

func (t *Tokens) ToBase64() (string, error) {
	jsonBytes, err := json.Marshal(t)
	if err != nil {
		return "", nil
	}

	base64Tokens := base64.StdEncoding.EncodeToString(jsonBytes)

	return base64Tokens, nil
}

func NewTokensFromBase64(tokenBase64 string) (*Tokens, error) {
	jsonBytes, err := base64.StdEncoding.DecodeString(tokenBase64)
	if err != nil {
		return nil, err
	}

	var tokenResponse Tokens
	err = json.Unmarshal(jsonBytes, &tokenResponse)
	if err != nil {
		return nil, err
	}

	return &tokenResponse, nil
}

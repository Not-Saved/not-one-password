package handler

import (
	"bytes"
	"context"
	"io"
	"main/internal/adapters/middleware"
	"main/internal/core/services"
	"main/internal/oapi"
)

type VaultHandler struct {
	vaultService *services.VaultService
}

func NewVaultHandler(vaultService *services.VaultService) *VaultHandler {
	return &VaultHandler{vaultService: vaultService}
}

func (h *VaultHandler) GetUserVault(ctx context.Context, request oapi.GetUserVaultRequestObject) (oapi.GetUserVaultResponseObject, error) {
	access, ok := middleware.GetAccessSession(ctx)
	if !ok || access == nil {
		return oapi.GetUserVault401JSONResponse{
			Code:    401,
			Message: "User not authenticated",
		}, nil
	}

	vault, err := h.vaultService.GetVaultByUserID(ctx, access.UserID)

	if err != nil {
		return oapi.GetUserVault500JSONResponse{
			InternalServerErrorJSONResponse: oapi.InternalServerErrorJSONResponse{
				Code:    500,
				Message: err.Error(),
			},
		}, nil
	}

	reader := bytes.NewReader(vault.Vault)
	contentLength := int64(len(vault.Vault))
	return oapi.GetUserVault200ApplicationoctetStreamResponse{
		Body:          reader, // this will be returned as application/octet-stream
		ContentLength: contentLength,
	}, nil
}

func (h *VaultHandler) InsertUserVault(ctx context.Context, request oapi.InsertUserVaultRequestObject) (oapi.InsertUserVaultResponseObject, error) {
	// Get the current user's ID from the access session
	access, ok := middleware.GetAccessSession(ctx)
	if !ok {
		return oapi.InsertUserVault401JSONResponse{
			Code:    401,
			Message: "User not authenticated",
		}, nil
	}

	// Read the raw binary vault from request body
	vaultBytes, err := io.ReadAll(request.Body)
	if err != nil {
		return oapi.InsertUserVault500JSONResponse{
			InternalServerErrorJSONResponse: oapi.InternalServerErrorJSONResponse{
				Code:    500,
				Message: "Failed to read request body",
			},
		}, nil
	}

	// Call the service to create or update the vault
	_, err = h.vaultService.InsertVaultByUserID(ctx, access.UserID, vaultBytes)
	if err != nil {
		return oapi.InsertUserVault500JSONResponse{
			InternalServerErrorJSONResponse: oapi.InternalServerErrorJSONResponse{
				Code:    500,
				Message: err.Error(),
			},
		}, nil
	}

	// Return 204 No Content
	return oapi.InsertUserVault204Response{}, nil
}

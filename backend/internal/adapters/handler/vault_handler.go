package handler

import (
	"context"
	"fmt"
	"io"
	"main/internal/adapters/middleware"
	"main/internal/core/services"
	"main/internal/oapi"
	"mime/multipart"
	"net/textproto"
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
	if vault == nil {
		return oapi.GetUserVault404JSONResponse{
			Code:    404,
			Message: "Vault not found",
		}, nil
	}

	return oapi.GetUserVault200MultipartResponse(func(writer *multipart.Writer) error {
		// --- Part 1: JSON metadata ---
		jsonHeader := textproto.MIMEHeader{}
		jsonHeader.Set("Content-Type", "application/json")
		jsonPart, err := writer.CreatePart(jsonHeader)
		if err != nil {
			return err
		}
		jsonContent := fmt.Sprintf(`{"updatedAt":"%d"}`, vault.UpdatedAt.Unix())
		if _, err := jsonPart.Write([]byte(jsonContent)); err != nil {
			return err
		}

		// --- Part 2: Binary vault ---
		binaryHeader := textproto.MIMEHeader{}
		binaryHeader.Set("Content-Type", "application/octet-stream")
		binaryPart, err := writer.CreatePart(binaryHeader)
		if err != nil {
			return err
		}
		if _, err := binaryPart.Write(vault.Vault); err != nil {
			return err
		}

		return nil
	}), nil
}

func (h *VaultHandler) PollUserVault(ctx context.Context, request oapi.PollUserVaultRequestObject) (oapi.PollUserVaultResponseObject, error) {
	access, ok := middleware.GetAccessSession(ctx)
	if !ok || access == nil {
		return oapi.PollUserVault401JSONResponse{
			Code:    401,
			Message: "User not authenticated",
		}, nil
	}

	vaultUpdateAt, err := h.vaultService.GetVaultUpdatedAtByUserID(ctx, access.UserID)

	if err != nil {
		return oapi.PollUserVault500JSONResponse{
			InternalServerErrorJSONResponse: oapi.InternalServerErrorJSONResponse{
				Code:    500,
				Message: err.Error(),
			},
		}, nil
	}
	if vaultUpdateAt == nil {
		return oapi.PollUserVault404JSONResponse{
			Code:    404,
			Message: "Vault not found",
		}, nil
	}

	return oapi.PollUserVault200JSONResponse{
		UpdatedAt: vaultUpdateAt.Unix(),
	}, nil
}

func (h *VaultHandler) InsertUserVault(ctx context.Context, request oapi.InsertUserVaultRequestObject) (oapi.InsertUserVaultResponseObject, error) {
	access, ok := middleware.GetAccessSession(ctx)
	if !ok {
		return oapi.InsertUserVault401JSONResponse{
			Code:    401,
			Message: "User not authenticated",
		}, nil
	}

	vaultBytes, err := io.ReadAll(request.Body)
	if err != nil {
		return oapi.InsertUserVault500JSONResponse{
			InternalServerErrorJSONResponse: oapi.InternalServerErrorJSONResponse{
				Code:    500,
				Message: "Failed to read request body",
			},
		}, nil
	}

	_, err = h.vaultService.InsertVaultByUserID(ctx, access.UserID, vaultBytes)
	if err != nil {
		return oapi.InsertUserVault500JSONResponse{
			InternalServerErrorJSONResponse: oapi.InternalServerErrorJSONResponse{
				Code:    500,
				Message: err.Error(),
			},
		}, nil
	}

	return oapi.InsertUserVault204Response{}, nil
}

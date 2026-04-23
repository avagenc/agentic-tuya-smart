package device

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/avagenc/zee-api/internal/domain"
	"github.com/go-chi/chi/v5"
	"go.naturallyfunny.dev/api"
	"go.naturallyfunny.dev/api/identity"
)

type Service interface {
	List(ctx context.Context, userID string) ([]domain.Device, error)
	SendCommands(ctx context.Context, userID string, deviceID string, commands []domain.DataPoint) (json.RawMessage, error)
}

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, err := identity.GetUserIDFromContext(r.Context())
	if err != nil {
		api.WriteError(w, api.NewError(api.Unauthenticated, "Missing user identity"))
		return
	}

	devices, err := h.svc.List(r.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrAccountNotLinked) {
			api.WriteError(w, api.NewError(api.Unauthenticated, "No Tuya App Account is linked to the user"))
			return
		}
		api.WriteError(w, api.NewError(api.Unavailable, err.Error()).WithError(err))
		return
	}

	api.WriteSuccess(w, api.OK, "Devices retrieved successfully", devices, nil)
}

func (h *Handler) SendCommands(w http.ResponseWriter, r *http.Request) {
	userID, err := identity.GetUserIDFromContext(r.Context())
	if err != nil {
		api.WriteError(w, api.NewError(api.Unauthenticated, "Missing user identity"))
		return
	}

	deviceID := chi.URLParam(r, "deviceId")
	if deviceID == "" {
		api.WriteError(w, api.NewError(api.InvalidArgument, "Missing deviceId"))
		return
	}

	var req struct {
		Commands []domain.DataPoint `json:"commands"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, api.NewError(api.InvalidArgument, "Invalid request body"))
		return
	}

	if len(req.Commands) == 0 {
		api.WriteError(w, api.NewError(api.InvalidArgument, "Commands cannot be empty"))
		return
	}

	result, err := h.svc.SendCommands(r.Context(), userID, deviceID, req.Commands)
	if err != nil {
		if errors.Is(err, domain.ErrAccountNotLinked) {
			api.WriteError(w, api.NewError(api.Unauthenticated, "No Tuya App Account is linked to the user"))
			return
		}
		if errors.Is(err, domain.ErrDeviceNotOwned) {
			api.WriteError(w, api.NewError(api.PermissionDenied, "Device does not belong to user"))
			return
		}
		api.WriteError(w, api.NewError(api.Unavailable, err.Error()).WithError(err))
		return
	}

	api.WriteSuccess(w, api.OK, "Commands sent successfully", result, nil)
}

package account

import (
	"context"
	"net/http"

	"github.com/avagenc/zee-api/pkg/api"
)

type Service interface {
	Get(ctx context.Context, ownerID string) (Account, error)
}

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	ownerID, err := api.GetUserIDFromContext(r.Context())
	if err != nil {
		api.WriteError(w, api.NewError(http.StatusUnauthorized, "UNAUTHORIZED", "Missing user identity"))
		return
	}

	acc, err := h.svc.Get(r.Context(), ownerID)
	if err != nil {
		api.WriteError(w, api.NewError(http.StatusNotFound, "NOT_FOUND", "Tuya account not linked"))
		return
	}

	api.WriteSuccess(w, http.StatusOK, "SUCCESS", "Account retrieved", map[string]any{
		"ownerId":   acc.OwnerID,
		"tuyaUid":   acc.TuyaUID,
		"createdAt": acc.CreatedAt,
		"updatedAt": acc.UpdatedAt,
	}, nil)
}

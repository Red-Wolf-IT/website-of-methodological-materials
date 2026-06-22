package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"website-of-methodological-materials/internal/models"
	"website-of-methodological-materials/internal/service"
	"website-of-methodological-materials/internal/validator"
)

type ManualHandler struct {
	service   *service.ManualService
	validator *validator.Validator
}

func NewManualHandler(service *service.ManualService, v *validator.Validator) *ManualHandler {
	return &ManualHandler{service: service, validator: v}
}

func (h *ManualHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateManualRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	if err := h.validator.Validate(req); err != nil {
		respondValidationError(w, validator.ToFieldErrors(err))
		return
	}

	manual, err := h.service.Create(r.Context(), req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create manual")
		return
	}

	respondData(w, http.StatusCreated, manual)
}

func (h *ManualHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid manual id")
		return
	}

	manual, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrManualNotFound) {
			respondError(w, http.StatusNotFound, "manual not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to get manual")
		return
	}

	respondData(w, http.StatusOK, manual)
}

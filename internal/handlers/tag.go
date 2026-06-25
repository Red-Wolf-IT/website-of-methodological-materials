package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"website-of-methodological-materials/internal/models"
	"website-of-methodological-materials/internal/service"
	"website-of-methodological-materials/internal/validator"
)

type TagHandler struct {
	service   *service.TagService
	validator *validator.Validator
}

func NewTagHandler(service *service.TagService, v *validator.Validator) *TagHandler {
	return &TagHandler{service: service, validator: v}
}

func (h *TagHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	if err := h.validator.Validate(req); err != nil {
		respondValidationError(w, validator.ToFieldErrors(err))
		return
	}

	tag, err := h.service.Create(r.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrTagNameTaken) {
			respondError(w, http.StatusConflict, "tag name already exists")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to create tag")
		return
	}

	respondData(w, http.StatusCreated, tag)
}

func (h *TagHandler) List(w http.ResponseWriter, r *http.Request) {
	tags, err := h.service.List(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list tags")
		return
	}

	respondData(w, http.StatusOK, tags)
}

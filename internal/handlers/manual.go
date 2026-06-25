package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"website-of-methodological-materials/internal/models"
	"website-of-methodological-materials/internal/service"
	"website-of-methodological-materials/internal/validator"
)

type ManualHandler struct {
	service   *service.ManualService
	validator *validator.Validator
	viewsChan chan<- uuid.UUID
}

func NewManualHandler(service *service.ManualService, v *validator.Validator, viewsChan chan<- uuid.UUID) *ManualHandler {
	return &ManualHandler{service: service, validator: v, viewsChan: viewsChan}
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

func (h *ManualHandler) List(w http.ResponseWriter, r *http.Request) {
	filter, err := parseManualListFilter(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.service.List(r.Context(), filter)
	if err != nil {
		if errors.Is(err, service.ErrInvalidListParams) {
			respondError(w, http.StatusBadRequest, "invalid query parameters")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to list manuals")
		return
	}

	respondData(w, http.StatusOK, result)
}

func parseManualListFilter(r *http.Request) (models.ManualListFilter, error) {
	q := r.URL.Query()
	filter := models.ManualListFilter{
		Author: strings.TrimSpace(q.Get("author")),
		Q:      strings.TrimSpace(q.Get("q")),
		Sort:   strings.TrimSpace(q.Get("sort")),
	}

	if tagIDStr := strings.TrimSpace(q.Get("tag_id")); tagIDStr != "" {
		tagID, err := strconv.Atoi(tagIDStr)
		if err != nil || tagID <= 0 {
			return filter, errors.New("invalid tag_id")
		}
		filter.TagID = &tagID
	}

	if pageStr := strings.TrimSpace(q.Get("page")); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			return filter, errors.New("invalid page")
		}
		filter.Page = page
	}

	if limitStr := strings.TrimSpace(q.Get("limit")); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 {
			return filter, errors.New("invalid limit")
		}
		filter.Limit = limit
	}

	return filter, nil
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

	select {
	case h.viewsChan <- id:
	default:
	}

	respondData(w, http.StatusOK, manual)
}

func (h *ManualHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid manual id")
		return
	}

	var req models.UpdateManualRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	if err := h.validator.Validate(req); err != nil {
		respondValidationError(w, validator.ToFieldErrors(err))
		return
	}

	manual, err := h.service.Update(r.Context(), id, req)
	if err != nil {
		if errors.Is(err, service.ErrManualNotFound) {
			respondError(w, http.StatusNotFound, "manual not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to update manual")
		return
	}

	respondData(w, http.StatusOK, manual)
}

func (h *ManualHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid manual id")
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		if errors.Is(err, service.ErrManualNotFound) {
			respondError(w, http.StatusNotFound, "manual not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to delete manual")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ManualHandler) UploadAttachment(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid manual id")
		return
	}

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		respondError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	manual, err := h.service.UploadAttachment(r.Context(), id, header.Filename, file)
	if err != nil {
		if errors.Is(err, service.ErrManualNotFound) {
			respondError(w, http.StatusNotFound, "manual not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to upload attachment")
		return
	}

	respondData(w, http.StatusOK, manual)
}

func (h *ManualHandler) AttachTags(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid manual id")
		return
	}

	var req models.AttachTagsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	if err := h.validator.Validate(req); err != nil {
		respondValidationError(w, validator.ToFieldErrors(err))
		return
	}

	manual, err := h.service.AttachTags(r.Context(), id, req.TagIDs)
	if err != nil {
		if errors.Is(err, service.ErrManualNotFound) {
			respondError(w, http.StatusNotFound, "manual not found")
			return
		}
		if errors.Is(err, service.ErrTagNotFound) {
			respondError(w, http.StatusNotFound, "tag not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to attach tags")
		return
	}

	respondData(w, http.StatusOK, manual)
}

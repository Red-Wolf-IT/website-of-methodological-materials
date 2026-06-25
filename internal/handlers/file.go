package handlers

import (
	"errors"
	"io"
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi/v5"

	"website-of-methodological-materials/internal/service"
)

const maxUploadSize = 10 << 20 // 10 MB

type FileHandler struct {
	service *service.ManualService
}

func NewFileHandler(service *service.ManualService) *FileHandler {
	return &FileHandler{service: service}
}

func (h *FileHandler) Serve(w http.ResponseWriter, r *http.Request) {
	filename := chi.URLParam(r, "filename")
	if filename == "" {
		respondError(w, http.StatusBadRequest, "invalid file name")
		return
	}

	webPath := "/uploads/" + filepath.Base(filename)

	file, err := h.service.OpenAttachment(webPath)
	if err != nil {
		if errors.Is(err, service.ErrFileNotFound) {
			respondError(w, http.StatusNotFound, "file not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to open file")
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "inline; filename=\""+filepath.Base(filename)+"\"")
	_, _ = io.Copy(w, file)
}

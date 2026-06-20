package handlers

import (
	"encoding/json"
	"net/http"

	"website-of-methodological-materials/internal/service"
)

type HealthHandler struct {
	service *service.HealthService
}

func NewHealthHandler(service *service.HealthService) *HealthHandler {
	return &HealthHandler{service: service}
}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	response := h.service.Check(r.Context())

	w.Header().Set("Content-Type", "application/json")
	// 503, если БД не отвечает — так ожидают оркестраторы и балансировщики
	if response.Status != "ok" {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	_ = json.NewEncoder(w).Encode(response)
}

package handlers

import (
	"encoding/json"
	"net/http"

	"website-of-methodological-materials/internal/validator"
)

type fieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type errorBody struct {
	Message string       `json:"message"`
	Fields  []fieldError `json:"fields,omitempty"`
}

type errorResponse struct {
	Error errorBody `json:"error"`
}

type dataResponse struct {
	Data any `json:"data"`
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func respondData(w http.ResponseWriter, status int, data any) {
	writeJSON(w, status, dataResponse{Data: data})
}

func RespondError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{
		Error: errorBody{Message: message},
	})
}

func respondError(w http.ResponseWriter, status int, message string) {
	RespondError(w, status, message)
}

func respondValidationError(w http.ResponseWriter, fields []validator.FieldError) {
	bodyFields := make([]fieldError, len(fields))
	for i, f := range fields {
		bodyFields[i] = fieldError{
			Field:   f.Field,
			Message: f.Message,
		}
	}

	writeJSON(w, http.StatusBadRequest, errorResponse{
		Error: errorBody{
			Message: "validation failed",
			Fields:  bodyFields,
		},
	})
}

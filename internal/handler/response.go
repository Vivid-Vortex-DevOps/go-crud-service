package handler

import (
	"encoding/json"
	"net/http"
)

type apiResponse struct {
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

type paginatedResponse struct {
	Data       any   `json:"data"`
	TotalCount int64 `json:"totalCount"`
	Page       int   `json:"page"`
	Size       int   `json:"size"`
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(apiResponse{Data: data}) //nolint:errcheck
}

func writePaginated(w http.ResponseWriter, data any, total int64, page, size int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(paginatedResponse{ //nolint:errcheck
		Data:       data,
		TotalCount: total,
		Page:       page,
		Size:       size,
	})
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(apiResponse{Error: message}) //nolint:errcheck
}

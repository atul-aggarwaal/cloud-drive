package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/atul-aggarwaal/cloud-drive/internal/domain"
	"github.com/atul-aggarwaal/cloud-drive/internal/usecase"
)

type FileShareHandler struct {
	service *usecase.FileShareService
}

func NewFileShareHandler(service *usecase.FileShareService) *FileShareHandler {
	return &FileShareHandler{service: service}
}

func (s *FileShareHandler) NewFileShareRequest(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userId, ok := r.Context().Value(UserIdKey).(string)
	if !ok || userId == "" {
		http.Error(w, "Unauthorized user", http.StatusUnauthorized)
		return
	}
	var fileShareRequest domain.FileShareRequest
	if err := json.NewDecoder(r.Body).Decode(&fileShareRequest); err != nil {
		http.Error(w, "Malformed request", http.StatusBadRequest)
		return
	}

	if err := s.service.ShareFile(r.Context(), userId, fileShareRequest); err != nil {
		http.Error(w, fmt.Sprintf("Error occurred while sharing file %w", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "File shared successfully"})
}

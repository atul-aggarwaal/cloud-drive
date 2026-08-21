package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/atul-aggarwaal/cloud-drive/internal/domain"
	"github.com/atul-aggarwaal/cloud-drive/internal/usecase"
)

type FileHandler struct {
	service *usecase.FileService
}

func NewFileHandler(service *usecase.FileService) *FileHandler {
	return &FileHandler{service: service}
}

/*
Sample request

	{
		"file_name": "vacation_vidoe.mp4",
		"size": 106324234
	}
*/
type InitiateUploadRequest struct {
	FileName string `json:"file_name"`
	Size     int64  `json:"size"`
	FileHash string `json:"file_hash"`
}

func (this *FileHandler) HandleInitiateUpload(writer http.ResponseWriter, request *http.Request) {
	//1. verify HTTP method (only POST allowed)
	if request.Method != http.MethodPost {
		http.Error(writer, domain.ErrorMethodNotAllowed.Error(), http.StatusMethodNotAllowed)
		return
	}

	userId, ok := request.Context().Value(UserIdKey).(string)
	if !ok || userId == "" {
		http.Error(writer, domain.ErrorUnauthorizedUser.Error(), http.StatusUnauthorized)
		return
	}

	//2. Decode incoming Json Pipe
	var req InitiateUploadRequest //Creates an instance of business object and maps incmoing request to this
	if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
		http.Error(writer, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 3. Validate input
	if req.FileName == "" || req.Size <= 0 || req.FileHash == "" {
		http.Error(writer, "Missing mandatory execution properties: file_name, size, and file_hash are required", http.StatusBadRequest)
		return
	}
	log.Printf("request validated successfully")
	//3. Hand-off the work to service
	file, uploadUrl, err := this.service.InitiateUpload(request.Context(), userId, req.FileName, req.FileHash, req.Size)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	//4. send back the success response
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusCreated) //201 Created

	/*
			Sample	response
		{
		  "file": {
		    "id": "550e8400-e29b-41d4-a716-446655440000",
		    "user_id": "rocky_02",
		    "file_name": "vacation_video.mp4",
		    "size": 1073741824,
		    "status": "PENDING",
		    "created_at": "2026-05-14T23:15:00Z"
		  },
		  "upload_url": "https://s3.localstack.localhost:4566/cloud-drive/..."
		}
	*/
	json.NewEncoder(writer).Encode(map[string]interface{}{
		"file":       file,
		"upload_url": uploadUrl,
	})
}

func (this *FileHandler) DownloadFile(writer http.ResponseWriter, request *http.Request) {
	log.Println("Handling Download File Request")
	//Make sure it is HTTP GET method
	if request.Method != http.MethodGet {
		http.Error(writer, domain.ErrorMethodNotAllowed.Error(), http.StatusMethodNotAllowed)
		return
	}

	userId := request.Context().Value(UserIdKey).(string)
	fileId := request.URL.Query().Get("file_id")

	log.Printf("file id: %s - user id: %s", fileId, userId)

	if fileId == "" || userId == "" {
		http.Error(writer, "Invalid request", http.StatusBadRequest)
		return
	}

	downloadUrl, err := this.service.InitiateDownload(request.Context(), fileId, userId)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(writer, request, downloadUrl, http.StatusFound)
}

func (this *FileHandler) ListFiles(writer http.ResponseWriter, request *http.Request) {

	if http.MethodGet != request.Method {
		http.Error(writer, "method not allowed ", http.StatusMethodNotAllowed)
	}

	userId, ok := request.Context().Value(UserIdKey).(string)

	if !ok || userId == "" {
		http.Error(writer, domain.ErrorUnauthorizedUser.Error(), http.StatusUnauthorized)
		return
	}
	files, error := this.service.ListFilesForUser(request.Context(), userId)

	if error != nil {
		log.Panic(error)
		http.Error(writer, "Error retrieving files for user", http.StatusInternalServerError)
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)

	json.NewEncoder(writer).Encode(files)
}

func (this *FileHandler) DeleteFile(writer http.ResponseWriter, request *http.Request) {
	if http.MethodDelete != request.Method {
		http.Error(writer, domain.ErrorMethodNotAllowed.Error(), http.StatusMethodNotAllowed)
		return
	}

	fileId := request.URL.Query().Get("id")
	if fileId == "" {
		http.Error(writer, "invalid request: missing file id", http.StatusBadRequest)
		return
	}

	userId, ok := request.Context().Value(UserIdKey).(string)
	if !ok || userId == "" {
		http.Error(writer, domain.ErrorUnauthorizedUser.Error(), http.StatusUnauthorized)
		return
	}

	err := this.service.RequestFileDeletion(request.Context(), userId, fileId)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}
package handler

import (
	"encoding/json"
	"log"
	"net/http"

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
		http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userId := request.Context().Value(UserIdKey).(string)
	if userId == "" {
		http.Error(writer, "Unauthorized", http.StatusUnauthorized)
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
	log.Println("Initiated Download File")
	//Make sure it is HTTP GET method
	if request.Method != http.MethodGet {
		http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	fileId := request.URL.Query().Get("file_id")
	userId := request.URL.Query().Get("ctx_verified_user_id")
	log.Println("File ID: %s - User ID: %s ", fileId, userId)

	if fileId == "" || userId == "" {
		http.Error(writer, "Invalid request", http.StatusBadRequest)
	}

	downloadUrl, err := this.service.InitiateDownload(request.Context(), fileId, userId)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(writer, request, downloadUrl, http.StatusFound)
}

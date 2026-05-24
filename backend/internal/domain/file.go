package domain

import (
	"time"
)

type User struct {
	ID           string    `json:"id"`
	UserName     string    `json:"user_name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` //Do no serialize passwords
	CreatedAt    time.Time `json:"created_at"`
}

type File struct {
	ID        string    `json:"id"`
	OwnerID   string    `json:"owner_id"`
	FileName  string    `json:"file_name"`
	IsFolder  bool      `json:"is_folder"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type FileVersion struct {
	ID         string `json:"id,omitempty"`
	FileId     string `json:"file_id"`
	VersionNum int    `json:"version_num"`
	FileHash   string `json:"file_hash"` //SHA-256
	Size       int64  `json:"size"`
	Status     string `json:"status"` //'PENDING','AVAILABLE','DELETED'
	CreatedAt  string `json:"created_at"`
}

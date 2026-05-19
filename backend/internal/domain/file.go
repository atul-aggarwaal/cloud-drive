package domain

import (
	"time"
)

type File struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	FileName  string    `json:"file_name"`
	Size      int64     `json:"size"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

package domain

import "time"

// Struct represents a single file share record
type FileShare struct {
	ID              string    `json:id`               //Unique Share ID
	FileID          string    `json:file_id`          // ID of the file being shared
	UserID          string    `json:user_id`          // ID of the user with whome file is being shared
	permissionLevel string    `json:permission_level` // Level of permission to the user "read" / "read_write"
	CreatedAt       time.Time `json:created_at`       // Time at which permission was given
}

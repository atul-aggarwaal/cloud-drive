package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/atul-aggarwaal/cloud-drive/internal/domain"
)

// PostgresFileRepository implements the domain.FileRepository interface
// using a PostgreSQL database.
type PostgresFileRepository struct {
	db *sql.DB
}

// NewPostresFileRepository creates a new instance of PostgresFileRepository.
func NewPostresFileRepository(db *sql.DB) *PostgresFileRepository {
	return &PostgresFileRepository{db: db}
}

// CreateFile inserts a new file metadata record into the database.
func (r *PostgresFileRepository) CreateFile(ctx context.Context, file *domain.File) error {
	query := `INSERT INTO files(id, owner_id, file_name, is_folder, created_at, updated_at) 
			VALUES($1, $2, $3, $4, NOW(), NOW())` //Set create and update time as now.
	_, err := r.db.ExecContext(ctx, query, file.ID, file.OwnerID, file.FileName, file.IsFolder)

	return err
}

// CreateVersion inserts a new file version record into the database.
func (r *PostgresFileRepository) CreateVersion(ctx context.Context, fileVersion *domain.FileVersion) error {
	query := `INSERT INTO file_versions( file_id, version_num, file_hash, size, status, created_at)
			VALUES ($1, $2, $3, $4, $5, NOW())` //version Id is auto increment bigserial

	_, err := r.db.ExecContext(ctx, query, fileVersion.FileId, fileVersion.VersionNum, fileVersion.FileHash, fileVersion.Size, fileVersion.Status)

	return err
}

// UpdateVersionStatus updates the upload status of a file.
func (r *PostgresFileRepository) UpdateVersionStatus(ctx context.Context, fileId string, versionNum int, status string) error {
	query := `UPDATE file_versions SET status = $1 WHERE file_id = $2 AND version_num = $3`

	_, err := r.db.ExecContext(ctx, query, status, fileId, versionNum)

	return err
}

// GetFileByID retrieves a file metadata record by its ID.
func (r *PostgresFileRepository) GetFileByID(ctx context.Context, id string) (*domain.File, error) {
	log.Printf("Fetching file metadata for ID: %s", id)
	query := `SELECT id, owner_id, file_name, is_folder, created_at, updated_at FROM files WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)

	var file domain.File
	err := row.Scan(&file.ID, &file.OwnerID, &file.FileName, &file.IsFolder, &file.CreatedAt, &file.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Println("Record not found $v:", err)
			return nil, nil
		}
		return nil, err
	}
	return &file, nil
}

// GetLatestVersion retrieves the latest file version record by its ID.
func (r *PostgresFileRepository) GetLatestVersion(ctx context.Context, fileId string) (*domain.FileVersion, error) {
	query := `SELECT id, file_id, version_num, file_hash, size, status, created_at FROM file_versions WHERE file_id =$1 ORDER BY id DESC LIMIT 1`
	row := r.db.QueryRowContext(ctx, query, fileId)

	var fileVersion domain.FileVersion
	err := row.Scan(&fileVersion.ID, &fileVersion.FileId, &fileVersion.VersionNum, &fileVersion.FileHash, &fileVersion.Size, &fileVersion.Status, &fileVersion.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Println("Record not found $v", err)
			return nil, err
		}
		return nil, err
	}

	return &fileVersion, nil
}

func(r *PostgresFileRepository) GetVersions(ctx context.Context, fileID string) ([]*domain.FileVersion, error){
	query := `SELECT id, file_id, version_num, file_hash, size, status, created_at FROM file_versions WHERE file_id = $1`
	rows,err := r.db.QueryContext(ctx, query, fileID)

	if err != nil{
		return nil, fmt.Errorf("query file versions: %w", err)
	}
	defer rows.Close()
	var fileVersions []*domain.FileVersion
	
	for rows.Next(){
		var fileVersion domain.FileVersion

		err := rows.Scan(
				&fileVersion.ID, 
				&fileVersion.FileId, 
				&fileVersion.VersionNum, 
				&fileVersion.FileHash, 
				&fileVersion.Size, 
				&fileVersion.Status, 
				&fileVersion.CreatedAt,
			)
		if err !=nil{
			return nil, fmt.Errorf("scanning file versions: %w", err)
		}
		fileVersions = append(fileVersions, &fileVersion)
	}

	if  err:= rows.Err(); err!=nil {
		return nil, fmt.Errorf("iterating rows error %w", err)
	}

	return fileVersions, nil
}

func (r *PostgresFileRepository) GetFiles(ctx context.Context, userId string) ([]*domain.File, error) {
	query := `SELECT id, owner_id, file_name, is_folder, created_at, updated_at 
				FROM files 
				WHERE 
					owner_id = $1
				AND deleted_at IS null
				OR EXISTS(
							SELECT 1 
							FROM file_shares 
							WHERE file_shares.file_id = files.id 
							AND file_shares.shared_with_user_id = $1
						)`
	
	rows, err := r.db.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []*domain.File
	for rows.Next() {
		var file domain.File
		err := rows.Scan(&file.ID, &file.OwnerID, &file.FileName, &file.IsFolder, &file.CreatedAt, &file.UpdatedAt)
		if err != nil {
			return nil, err
		}
		files = append(files, &file)
	}
	
	return files, nil
}

func(r *PostgresFileRepository) SoftDeleteFileMetadata(ctx context.Context, fileId string) error{
	query := `UPDATE files SET deleted_at = $1 WHERE id = $2`

	_, err := r.db.ExecContext(ctx, query,time.Now(), fileId)
	

	if err!=nil{
		return fmt.Errorf("soft delete failed:  %w",err)
	}
	return nil
}

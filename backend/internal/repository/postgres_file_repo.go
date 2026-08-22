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

//creates initial file and its first version in a single transaction. This is to ensure that both file and its first version are created together or none at all.
func (r *PostgresFileRepository) CreateFileWithInitialVersion(ctx context.Context, file *domain.File, fileVersion *domain.FileVersion) error {

	tx, err :=r.db.BeginTx(ctx,nil)

	if err!=nil{
		return fmt.Errorf("starting db transaction: %w",err)
	}

	err = r.CreateFile(ctx, tx, file)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("creating file: %w", err)
	}

	err = r.CreateVersion(ctx, tx, fileVersion)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("creating file version: %w", err)
	}

	return tx.Commit()
}

// CreateFile inserts a new file metadata record into the database.
func (r *PostgresFileRepository) CreateFile(ctx context.Context, tx *sql.Tx, file *domain.File) error {
	query := `INSERT INTO files(id, owner_id, file_name, is_folder, created_at, updated_at) 
			VALUES($1, $2, $3, $4, NOW(), NOW())` //Set create and update time as now.
	_, err := tx.ExecContext(ctx, query, file.ID, file.OwnerID, file.FileName, file.IsFolder)

	return err
}

// CreateVersion inserts a new file version record into the database.
func (r *PostgresFileRepository) CreateVersion(ctx context.Context, tx *sql.Tx, fileVersion *domain.FileVersion) error {
	query := `INSERT INTO file_versions( file_id, version_num, file_hash, size, status, created_at)
			VALUES ($1, $2, $3, $4, $5, NOW())` //version Id is auto increment bigserial

	_, err := tx.ExecContext(ctx, query, fileVersion.FileId, fileVersion.VersionNum, fileVersion.FileHash, fileVersion.Size, fileVersion.Status)

	return err
}

// UpdateVersionStatus updates the upload status of a file version.
func (r *PostgresFileRepository) UpdateFileVersionStatus(ctx context.Context, fileId string, versionNum int, expectedStatus string, newStatus string) error {
	query := `UPDATE file_versions 
			SET status = $1 
			WHERE file_id = $2 
			AND version_num = $3
			AND status = $4`

	result, err := r.db.ExecContext(ctx, query, newStatus, fileId, versionNum, expectedStatus)
	if err != nil {
		return err
	}

	rowsAffected, err :=result.RowsAffected()
	if err != nil {
		return fmt.Errorf("updating version status failed: %w", err)
	}

	//No File record matched with current requirement of status, version and file id. Thus, no transition happened.
	if rowsAffected == 0{
		log.Printf("No rows updated for fileId=%s, versionNum=%d, expectedStatus=%s, newStatus=%s", fileId, versionNum, expectedStatus, newStatus)
	}
	return nil
}
// updates the upload status of a file.
func (r *PostgresFileRepository) UpdateFileStatus(ctx context.Context, fileId string, expectedStatus string, newStatus string) error {
	query := `UPDATE files 
			SET status = $1 
			WHERE file_id = $2 
			AND status = $4`

	result, err := r.db.ExecContext(ctx, query, newStatus, fileId, expectedStatus)
	if err != nil {
		return err
	}

	rowsAffected, err :=result.RowsAffected()
	if err != nil {
		return fmt.Errorf("updating file status failed: %w", err)
	}

	//No File record matched with current requirement of status, version and file id. Thus, no transition happened.
	if rowsAffected == 0{
		log.Printf("No rows updated for fileId=%s,expectedStatus=%s, newStatus=%s", fileId, expectedStatus, newStatus)
	}
	return nil
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
			return nil, domain.ErrorFileNotFound
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

func (r *PostgresFileRepository) GetVersions(ctx context.Context, fileID string) ([]*domain.FileVersion, error) {
	query := `SELECT id, file_id, version_num, file_hash, size, status, created_at FROM file_versions WHERE file_id = $1`
	rows, err := r.db.QueryContext(ctx, query, fileID)

	if err != nil {
		return nil, fmt.Errorf("query file versions: %w", err)
	}
	defer rows.Close()
	var fileVersions []*domain.FileVersion

	for rows.Next() {
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
		if err != nil {
			return nil, fmt.Errorf("scanning file versions: %w", err)
		}
		fileVersions = append(fileVersions, &fileVersion)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating rows error %w", err)
	}
	return fileVersions, nil
}

func (r *PostgresFileRepository) GetFiles(ctx context.Context, userId string) ([]*domain.File, error) {
	query := `SELECT id, owner_id, file_name, is_folder, created_at, updated_at 
				FROM files 
				WHERE 
					owner_id = $1
				AND delete_requested_at IS null
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

	// catch errors if any, occurred during iteration
	if err := rows.Err(); err != nil {
		return nil, err
	}
	rows.Close()
	return files, nil
}

func (r *PostgresFileRepository) RequestDelete(ctx context.Context, userId string, fileId string) error {
	query := `UPDATE files SET 
								delete_requested_at = $1, 
								deleted_requested_by = $2,
								lifecycle_status = $3,
								updated_by = $4
				WHERE id = $5`

	_, err := r.db.ExecContext(ctx, query, time.Now(), userId, domain.FileStatusDeleteRequested, userId, fileId)

	if err != nil {
		return fmt.Errorf("soft delete failed:  %w", err)
	}
	return nil
}

/*
*

	Retrieve files which are marked for deletion. Pick 50 files at a time and lock them for processing.
	This is to avoid multiple workers picking up the same file for deletion and skip files which are already locked by another worker.
*/
func (r *PostgresFileRepository) GetFilesMarkedForDeletion(ctx context.Context) ([]*domain.File, error) {
	query := `SELECT
					id,
					owner_id,
					file_name,
					is_folder,
					created_at,
					updated_at
				FROM files
				WHERE lifecycle_status = $1
				ORDER BY created_at
				LIMIT 50
				FOR UPDATE SKIP LOCKED`

	rows, err := r.db.QueryContext(ctx, query, domain.FileStatusDeleteRequested)

	if err != nil {
		return nil, err
	}
	var files []*domain.File
	defer rows.Close()
	for rows.Next() {
		var file domain.File
		err := rows.Scan(&file.ID, &file.OwnerID, &file.FileName, &file.IsFolder, &file.CreatedAt, &file.UpdatedAt)
		if err != nil {
			return nil, err
		}
		files = append(files, &file)
	}

	// catch errors if any, occurred during iteration
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return files, nil
}

/*
*

	Retrieve files which are marked for deletion. Pick 50 files at a time and lock them for processing.
	This is to avoid multiple workers picking up the same file for deletion and skip files which are already locked by another worker.
*/
func (r *PostgresFileRepository) ClaimFilesForDeletion(ctx context.Context, limit int) ([]*domain.File, error) {

	transaction, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("starting transaction: %w", err)
	}

	defer transaction.Rollback() // Rollback the transaction if not committed

	query := `WITH claimed AS (
					SELECT
						id,
						owner_id,
						file_name,
						is_folder,
						created_at,
						updated_at
					FROM files
					WHERE lifecycle_status = $1
					ORDER BY created_at
					LIMIT $2
					FOR UPDATE SKIP LOCKED
				)
				UPDATE files f
				SET lifecycle_status = $3,
					updated_at = NOW()
				FROM claimed c
				where f.id = c.id
				RETURNING f.*`

	rows, err := transaction.QueryContext(ctx, query, domain.FileStatusDeleteRequested, limit,domain.FileStatusDeleting)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var files []*domain.File
	for rows.Next() {
		var file domain.File
		err := rows.Scan(
			&file.ID,
			&file.OwnerID,
			&file.FileName,
			&file.IsFolder,
			&file.CreatedAt,
			&file.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		files = append(files, &file)
	}
	// catch errors if any, occurred during iteration
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := transaction.Commit(); err != nil {
		return nil, fmt.Errorf("committing transaction: %w", err)
	}

	return files, nil
}

func (r *PostgresFileRepository) DeleteFileVersion(ctx context.Context, versionNum int, fileID string) error {
	query := `DELETE FROM file_versions 
					WHERE 
						 version_num =$1
					AND	 file_id =$2`

	result, err := r.db.ExecContext(ctx, query, versionNum, fileID)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("soft delete failed %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("soft delete failed: no file %s with version num %d found", fileID, versionNum)
	}
	return nil
}

func (r *PostgresFileRepository) MarkFileDeleted(ctx context.Context, fileID string) error {
	query := `UPDATE files SET 
								updated_at = $1,
								lifecycle_status = $2
				WHERE id = $3`

	result, err := r.db.ExecContext(ctx, query, time.Now(), domain.FileStatusDeleted, fileID)

	if err != nil {
		return fmt.Errorf("soft delete failed:  %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("soft delete failed %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("soft delete failed: no file found with id=%s", fileID)
	}

	return nil
}

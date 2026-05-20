package repository

import (
	"context"
	"database/sql"
	"errors"
	"log"

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

// Save inserts a new file record into the database.
func (r *PostgresFileRepository) Save(ctx context.Context, file *domain.File) error {
	query := `INSERT INTO files(id, user_id, file_name, size, status, created_at)
		VALUES($1, $2, $3, $4, $5, $6)`
	_, err := r.db.ExecContext(ctx, query, file.ID, file.UserID, file.FileName, file.Size, file.Status, file.CreatedAt)

	return err
}

// UpdateStatus updates the upload status of a file.
func (r *PostgresFileRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	query := `UPDATE files SET status = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, status, id)

	return err
}

// GetByID retrieves a file record by its ID from the database.
// Returns nil, nil if the file is not found.
func (r *PostgresFileRepository) GetByID(ctx context.Context, id string) (*domain.File, error) {
	query := `SELECT id, user_id, file_name, size, status, created_at from files WHERE id = $1`

	row := r.db.QueryRowContext(ctx, query, id)

	var f domain.File
	err := row.Scan(&f.ID, &f.UserID, &f.FileName, &f.Size, &f.Status, &f.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Println("Record not found $v:", err)
			return nil, nil //Not an application error, Just means "not found"
		}
		return nil, err
	}
	return &f, nil
}

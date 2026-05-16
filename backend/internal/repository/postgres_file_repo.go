package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/atul-aggarwaal/cloud-drive/internal/domain"
)

// PotgresFileRepository implements the domain.FileRepository interface
type PostgresFileRepository struct {
	db *sql.DB
}

// NewPostgresFileRepository is the constructor
func NewPostresFileRepository(db *sql.DB) *PostgresFileRepository {
	return &PostgresFileRepository{db: db}
}

// Save inserts a new file record into the database
func (this *PostgresFileRepository) Save(ctx context.Context, file *domain.File) error {
	query := `INSERT INTO files(id, user_id, file_name, size, status, created_at)
		VALUES($1, $2, $3, $4, $5, $6)`
	_, err := this.db.ExecContext(ctx, query, file.ID, file.UserID, file.FileName, file.Size, file.Status, file.CreatedAt)

	return err
}

// GetByID retrieves a file record by its ID
func (this *PostgresFileRepository) GetByID(ctx context.Context, id string) (*domain.File, error) {
	query := `SELECT id, user_id, file_name, size, status, created_at from files WHERE id = $1`

	row := this.db.QueryRowContext(ctx, query, id)

	var f domain.File
	err := row.Scan(&f.ID, &f.UserID, &f.FileName, &f.Size, &f.Status, &f.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil //Not an application error, Just means "not found"
		}
		return nil, err
	}
	return &f, nil
}

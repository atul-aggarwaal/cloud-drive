package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/atul-aggarwaal/cloud-drive/internal/domain"
)

var _ domain.FileShareRepository = (*PostgresFileShareRepository)(nil)

type PostgresFileShareRepository struct {
	db *sql.DB
}

func NewPostgresFileShareRepository(db *sql.DB) *PostgresFileShareRepository {
	return &PostgresFileShareRepository{db: db}
}

func (r *PostgresFileShareRepository) CreateShare(ctx context.Context, share *domain.FileShare) error {
	query := `INSERT INTO file_shares(id, file_id, shared_with_user_id, permission_level, created_at)
	VALUES($1, $2, $3, $4, NOW()) 
	ON CONFLICT(file_id, shared_with_user_id)
	DO UPDATE SET permission_level= EXCLUDED.permission_level`

	_, err := r.db.ExecContext(ctx, query, share.ID, share.FileID, share.SharedWithUserID, share.PermissionLevel)
	if err != nil {
		return fmt.Errorf("failed to update file share defails %w", err)
	}

	return nil
}

func (r *PostgresFileShareRepository) HasAccess(ctx context.Context, fileId string, userId string) (bool, error) {
	query := `SELECT EXISTS(
			SELECT 1 FROM file_shares WHERE file_id = $1 AND shared_with_user_id = $2
			UNION ALL 
			SELECT 1 FROM files WHERE id = $1 AND  owner_id = $2
			)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, fileId, userId).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("access rights validation failed %w", err)
	}

	return exists, nil
}

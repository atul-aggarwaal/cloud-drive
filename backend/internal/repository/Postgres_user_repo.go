package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/atul-aggarwaal/cloud-drive/internal/domain"
)

type PostresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) *PostresUserRepository {
	return &PostresUserRepository{db: db}
}

func (p PostresUserRepository) CreateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	query := `INSERT INTO users(id, username, email, password_hash, created_at)
				VALUES($1, $2, $3, $4, NOW())`
	_, err := p.db.ExecContext(ctx, query, user.ID, user.UserName, user.Email, user.PasswordHash)

	if err != nil {
		return nil, fmt.Errorf("user creation failed :%w", err)
	}
	return user, nil
}

func (p PostresUserRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT id, username, email, password_hash, created_at FROM users WHERE email=$1`
	row := p.db.QueryRowContext(ctx, query, email)

	var user domain.User
	err := row.Scan(&user.ID, &user.UserName, &user.Email, &user.PasswordHash, &user.CreatedAt)
	log.Printf("User Found: %+v", user)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// TODO : Add uniqueness to userName, this can be login ID as well in future.
func (p PostresUserRepository) GetUserByName(ctx context.Context, name string) (*domain.User, error) {
	log.Printf("postgres_user_repo__getUserByName")
	query := `SELECT id, username, email, password_hash, created_at FROM users WHERE username=$1`
	row := p.db.QueryRowContext(ctx, query, name)

	var user domain.User
	err := row.Scan(&user.ID, &user.UserName, &user.Email, &user.PasswordHash, &user.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (p PostresUserRepository) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	query := `SELECT id, username, email, password_hash, created_at FROM users WHERE id=$1`
	row := p.db.QueryRowContext(ctx, query, userID)

	var user domain.User
	err := row.Scan(&user.ID, &user.UserName, &user.Email, &user.PasswordHash, &user.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

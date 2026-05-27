package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/atul-aggarwaal/cloud-drive/internal/domain"
	"github.com/atul-aggarwaal/cloud-drive/internal/pkg/crypto"
)

type UserService struct {
	repo domain.UserRepository
}

func NewUserService(userRepo domain.UserRepository) *UserService {
	return &UserService{repo: userRepo}
}

func (this *UserService) registerUser(ctx context.Context, name string, email string, password string) (string, error) {

	//Validate Email ID
	existingUserByEmail, err := this.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", fmt.Errorf("Email validation failed: %w", err)
	}
	if existingUserByEmail != nil {
		return "", fmt.Errorf("Email already exists")
	}

	// Validate User Name
	existingUserByName, err := this.repo.GetUserByName(ctx, name)
	if err != nil {
		return "", fmt.Errorf("Name validation failed: %w", err)
	}
	if existingUserByName != nil {
		return "", fmt.Errorf("Username already exists")
	}

	passwordHashed, err := crypto.HashedPassword(password)
	if err != nil {
		return "", fmt.Errorf("Password hashing failed: %w", err)
	}
	userId, err := crypto.GenerateUUID7()
	if err != nil {
		return "", fmt.Errorf("User ID generation failed: %w", err)
	}

	newUser := &domain.User{
		ID:           userId,
		UserName:     name,
		Email:        email,
		PasswordHash: passwordHashed,
		CreatedAt:    time.Time{},
	}

	userName, err := this.repo.CreateUser(ctx, newUser)
	if err != nil {
		return "", fmt.Errorf("User Creation failed: %w", err)
	}
	return userName, nil
}

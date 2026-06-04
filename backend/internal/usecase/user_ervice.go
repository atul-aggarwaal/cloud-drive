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

func (this *UserService) RegisterUser(ctx context.Context, name string, email string, password string) (*domain.User, error) {

	//Validate Email ID
	existingUserByEmail, err := this.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("Email validation failed: %w", err)
	}
	if existingUserByEmail != nil {
		return nil, fmt.Errorf("Email already exists")
	}

	// Validate User Name
	existingUserByName, err := this.repo.GetUserByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("Name validation failed: %w", err)
	}
	if existingUserByName != nil {
		return nil, fmt.Errorf("Username already exists")
	}

	passwordHashed, err := crypto.HashedPassword(password)
	if err != nil {
		return nil, fmt.Errorf("Password hashing failed: %w", err)
	}
	userId, err := crypto.GenerateUUID7()
	if err != nil {
		return nil, fmt.Errorf("User ID generation failed: %w", err)
	}

	newUser := &domain.User{
		ID:           userId,
		UserName:     name,
		Email:        email,
		PasswordHash: passwordHashed,
		CreatedAt:    time.Time{},
	}

	user, err := this.repo.CreateUser(ctx, newUser)
	if err != nil {
		return nil, fmt.Errorf("User Creation failed: %w", err)
	}
	return user, nil
}

// Authenticate User based with email and password
func (this *UserService) AuthenticateUser(ctx context.Context, email string, password string) (*domain.User, error) {
	user, err := this.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: Invalid email/password")
	}

	if user == nil {
		return nil, fmt.Errorf("authentication failed: user doesn't exist")
	}

	passwordMatch := crypto.VerifyPassword(password, user.PasswordHash)
	if passwordMatch {
		return nil, fmt.Errorf("authentication failed: Invalid email/password")
	}

	return user, nil
}

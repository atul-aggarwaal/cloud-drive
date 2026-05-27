package domain

import "context"

// UserRepository defines the contract for User related operations
type UserRepository interface {

	//Creates a new User in the system
	CreateUser(ctx context.Context, user *User) (*User, error)
	//Retrieves the user details using User's email ID
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	//Retrieves User details using username, which is always unique
	GetUserByName(ctx context.Context, name string) (*User, error)
	//Retrieves User details using User's name
	GetUserByID(ctx context.Context, name string) (*User, error)
}

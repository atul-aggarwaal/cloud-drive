package crypto

import (
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Generates a Time sorted UUID v7
func GenerateUUID7() (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("failed to generate UUID: %w", err)
	}
	return id.String(), nil
}

// Encryptts plane text password into irreversible and cryptographically salted footprint
func HashedPassword(plainText string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(plainText), 13)
	if err != nil {
		return "", fmt.Errorf("failed to generate hashed password: %w", err)
	}
	return string(hashedBytes), nil
}

// Hash plain text password using salt and cost embedded in hashed password and compare
// output with hashed password, If output is same, password match is success else fail.
func VerifyPassword(password string, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err != nil
}

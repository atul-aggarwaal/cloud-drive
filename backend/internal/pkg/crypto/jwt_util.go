package crypto

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecretKey = []byte("This_is_a_secure_key_for_JWT_ideally_it_should_Be_private_key_generated_using_128_bit_algorithm")

type UserClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"user_name"`
	jwt.RegisteredClaims
}

// Generates a new JWT Access Token for given User ID and User Name
func GenerateAccessToken(userID string, userName string) (string, error) {
	claims := UserClaims{
		UserID:   userID,
		Username: userName,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "cloud-drive-service",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(jwtSecretKey)

	if err != nil {
		return "", fmt.Errorf("failed to issue new token : %w", err)
	}

	return signedToken, nil
}

// Validate incoming JWT token
func ValidateToken(tokenStr string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected cryptographic signature method: %v", token.Header["alg"])
		}

		return jwtSecretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token, token validation failed %w", err)
	}

	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

package token

import (
	"golang.org/x/crypto/bcrypt"
)

func ComputeToken(radiusSecret string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(radiusSecret), bcrypt.MinCost)
	return string(hash), err
}

func ValidateToken(radiusSecret string, token string) error {
	hash := []byte(token)
	password := []byte(radiusSecret)
	return bcrypt.CompareHashAndPassword(hash, password)
}
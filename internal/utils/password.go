package utils

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashedPassword returns a stringify hash of the password.
func HashedPassword(password string) (string, error) {
	buf := []byte(password)
	if len(buf) > 72 {
		fmt.Print(len(buf))
		return "", fmt.Errorf("password exceeded maximum length of 72")
	}

	hash, err := bcrypt.GenerateFromPassword(buf, bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("couldn't hash password: %w", err)
	}
	return string(hash), nil
}

// VerifyPassword checks the hashed password against
// the cleartext password.
func VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

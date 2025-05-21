package helpers

import (
	"encoding/base64"
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func HashPasswordBase64(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return "", fmt.Errorf("error hashing password: %w", err)
	}

	encodedHash := base64.StdEncoding.EncodeToString(bytes)

	return encodedHash, nil
}

func ComparePasswordBase64(hashedPasswordBase64 string, password string) error {
	hashedPassword, err := base64.StdEncoding.DecodeString(hashedPasswordBase64)

	if err != nil {
		return fmt.Errorf("error decoding hashed password: %w", err)
	}

	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))

	if err != nil {
		return fmt.Errorf("error comparing password: %w", err)
	}

	return nil
}

package main

import (
	"crypto/sha256"
	"encoding/base64"
	db "github.com/trenchesdeveloper/csv-reporter/db/sqlc"
	"github.com/trenchesdeveloper/csv-reporter/helpers"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"
)

type SignupRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type SigninRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

func (s *server) SignupHandler(w http.ResponseWriter, r *http.Request) {
	var req SignupRequest
	if err := readJSON(w, r, &req); err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := Validate.Struct(req); err != nil {
		errorResponse(w, http.StatusBadRequest, formatValidationErrors(err))
		return
	}

	// check if a user exists
	_, err := s.store.FindUserByEmail(r.Context(), req.Email)

	if err == nil {
		errorResponse(w, http.StatusConflict, "User already exists")
		return
	}

	hashedPassword, err := helpers.HashPasswordBase64(req.Password)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "Error hashing password")
		return
	}

	_, err = s.store.CreateUser(r.Context(), db.CreateUserParams{
		Email:          req.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		// log the error
		s.logger.Error("Error creating user", err)
		errorResponse(w, http.StatusInternalServerError, "Error creating user")
		return
	}

	jsonResponse(w, http.StatusCreated, nil, "User created successfully")
}

func (s *server) SigninHandler(w http.ResponseWriter, r *http.Request) {

	var req SigninRequest
	if err := readJSON(w, r, &req); err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := Validate.Struct(req); err != nil {
		errorResponse(w, http.StatusBadRequest, formatValidationErrors(err))
		return
	}

	user, err := s.store.FindUserByEmail(r.Context(), req.Email)
	if err != nil {
		errorResponse(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}
	if err := helpers.ComparePasswordBase64(user.HashedPassword, req.Password); err != nil {
		errorResponse(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// Generate JWT token
	token, err := s.tokenManager.GenerateTokenPairs(user.ID)

	if err != nil {
		s.logger.Error("Error generating token", err)
		errorResponse(w, http.StatusInternalServerError, "Error generating token")
		return
	}

	// delete the old refresh token
	err = s.store.DeleteAllUserRefreshTokens(r.Context(), user.ID)
	if err != nil {
		s.logger.Error("Error deleting old refresh tokens", err)
		errorResponse(w, http.StatusInternalServerError, "Error deleting old refresh tokens")
		return
	}
	// convert hashed token to base64
	hashedToken, err := hashToken(token.RefreshToken)
	// create refresh_token
	_, err = s.store.CreateRefreshToken(r.Context(), db.CreateRefreshTokenParams{
		HashedToken: hashedToken,
		UserID:      user.ID,
		ExpiresAt:   time.Now().Add(24 * time.Hour * 7),
	})

	if err != nil {
		s.logger.Error("Error creating refresh token", err)
		errorResponse(w, http.StatusInternalServerError, "Error creating refresh token")
		return
	}

	jsonResponse(w, http.StatusOK, token, "Signin successful")
}

func hashToken(plain string) (string, error) {
	// 1) Pre-hash:
	sum := sha256.Sum256([]byte(plain))

	// 2) Bcrypt the 32-byte digest
	bts, err := bcrypt.GenerateFromPassword(sum[:], bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	// 3) Base64 for safe storage
	return base64.StdEncoding.EncodeToString(bts), nil
}

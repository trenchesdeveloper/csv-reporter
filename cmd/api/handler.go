package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"github.com/google/uuid"
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

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
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
	if err != nil {
		s.logger.Error("Error hashing token", err)
		errorResponse(w, http.StatusInternalServerError, "Error hashing token")
		return
	}
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

func (s *server) RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest
	if err := readJSON(w, r, &req); err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := Validate.Struct(req); err != nil {
		errorResponse(w, http.StatusBadRequest, formatValidationErrors(err))
		return
	}

	// parse the token with tokenManager
	claims, err := s.tokenManager.ValidateToken(req.RefreshToken)
	if err != nil {
		s.logger.Error("Error validating token", err)
		errorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// get the subject from the claims
	userIdStr, err := claims.Claims.GetSubject()
	if err != nil {
		s.logger.Error("Error getting subject from claims", err)
		errorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	// check if the user ID is valid uuid
	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		s.logger.Error("Error parsing user ID", err)
		errorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// hash the refresh token
	hashedToken, err := hashToken(req.RefreshToken)
	if err != nil {
		s.logger.Error("Error hashing token", err)
		errorResponse(w, http.StatusInternalServerError, "Error hashing token")
		return
	}

	// query the database for the refresh token
	currentRefreshToken, err := s.store.GetTokenByPrimaryKey(r.Context(), userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.Error("Refresh token not found", err)
			errorResponse(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
		s.logger.Error("Error getting refresh token", err)
		errorResponse(w, http.StatusInternalServerError, "Error getting refresh token")
	}

	// check if the refresh token is expired
	if currentRefreshToken.ExpiresAt.Before(time.Now()) {
		s.logger.Error("Refresh token expired", err)
		errorResponse(w, http.StatusUnauthorized, "Refresh token expired")
		return
	}

	// create a new token
	token, err := s.tokenManager.GenerateTokenPairs(userId)
	if err != nil {
		s.logger.Error("Error generating token", err)
		errorResponse(w, http.StatusInternalServerError, "Error generating token")
		return
	}

	// delete the old refresh token
	err = s.store.DeleteAllUserRefreshTokens(r.Context(), userId)

	if err != nil {
		s.logger.Error("Error deleting old refresh tokens", err)
		errorResponse(w, http.StatusInternalServerError, "Error deleting old refresh tokens")
		return
	}

	// convert hashed token to base64
	hashedToken, err = hashToken(token.RefreshToken)

	if err != nil {
		s.logger.Error("Error hashing token", err)
		errorResponse(w, http.StatusInternalServerError, "Error hashing token")
		return
	}
	// create refresh_token
	_, err = s.store.CreateRefreshToken(r.Context(), db.CreateRefreshTokenParams{
		HashedToken: hashedToken,
		UserID:      userId,
		ExpiresAt:   time.Now().Add(24 * time.Hour * 7),
	})
	if err != nil {
		s.logger.Error("Error creating refresh token", err)
		errorResponse(w, http.StatusInternalServerError, "Error creating refresh token")
		return
	}
	// check if the user exists in the database
	jsonResponse(w, http.StatusOK, token, "Refresh token successful")
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

package main

import (
	db "github.com/trenchesdeveloper/csv-reporter/db/sqlc"
	"github.com/trenchesdeveloper/csv-reporter/helpers"
	"net/http"
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

	user, err := s.store.CreateUser(r.Context(), db.CreateUserParams{
		Email:          req.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		// log the error
		s.logger.Error("Error creating user", err)
		errorResponse(w, http.StatusInternalServerError, "Error creating user")
		return
	}

	jsonResponse(w, http.StatusCreated, user, "User created successfully")
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
	jsonResponse(w, http.StatusOK, token, "Signin successful")
}

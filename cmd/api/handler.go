package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-chi/chi/v5"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/google/uuid"
	db "github.com/trenchesdeveloper/csv-reporter/db/sqlc"
	"github.com/trenchesdeveloper/csv-reporter/helpers"
	"github.com/trenchesdeveloper/csv-reporter/reports"
	"golang.org/x/crypto/bcrypt"
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

type CreateReportRequest struct {
	ReportType string `json:"report_type" validate:"required,oneof=monsters weapons armor"`
}

type ReportResponse struct {
	ID                   uuid.UUID `json:"id"`
	ReportType           string    `json:"report_type,omitempty"`
	OutputFilePath       string    `json:"output_file_path,omitempty"`
	DownloadURL          string    `json:"download_url,omitempty"`
	DownloadUrlExpiresAt time.Time `json:"download_url_expires_at,omitempty"`
	StartedAt            time.Time `json:"started_at,omitempty"`
	CompletedAt          time.Time `json:"completed_at,omitempty"`
	FailedAt             time.Time `json:"failed_at,omitempty"`
	CreatedAt            time.Time `json:"created_at,omitempty"`
	ErrorMessage         string    `json:"error_message,omitempty"`
	Status               string    `json:"status"`
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

func (s *server) CreateReportHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateReportRequest
	if err := readJSON(w, r, &req); err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := Validate.Struct(req); err != nil {
		errorResponse(w, http.StatusBadRequest, formatValidationErrors(err))
		return
	}

	user, ok := UserFromContext(r)
	if !ok {
		errorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	report, err := s.store.CreateReport(r.Context(), db.CreateReportParams{
		UserID:     user.ID,
		ReportType: req.ReportType,
	})

	if err != nil {
		s.logger.Error("Error creating report", err)
		errorResponse(w, http.StatusInternalServerError, "Error creating report")
		return
	}

	//  send sqs message to build the report
	sqsMessage := reports.SQSMessage{
		UserID:   report.UserID,
		ReportID: report.ID,
	}
	queueUrl, err := s.sqsClient.GetQueueUrl(r.Context(), &sqs.GetQueueUrlInput{
		QueueName: aws.String(s.config.SQS_QUEUE),
	})

	if err != nil {
		s.logger.Error("Error getting SQS queue URL", err)
		errorResponse(w, http.StatusInternalServerError, "Error getting SQS queue URL")
		return
	}

	bytes, err := json.Marshal(sqsMessage)

	if err != nil {
		s.logger.Error("Error getting SQS queue URL", err)
		errorResponse(w, http.StatusInternalServerError, "Error getting SQS queue URL")
		return
	}

	_, err = s.sqsClient.SendMessage(r.Context(), &sqs.SendMessageInput{
		QueueUrl:    queueUrl.QueueUrl,
		MessageBody: aws.String(string(bytes)),
	})
	if err != nil {
		s.logger.Error("Error sending message to SQS", err)
		errorResponse(w, http.StatusInternalServerError, "Error sending message to SQS")
		return
	}

	reportResponse := ReportResponse{
		ID:                   report.ID,
		ReportType:           req.ReportType,
		StartedAt:            report.StartedAt.Time,
		Status:               GetStatus(report),
		OutputFilePath:       report.OutputFilePath.String,
		DownloadURL:          report.DownloadUrl.String,
		DownloadUrlExpiresAt: report.DownloadExpiresAt.Time,
		CompletedAt:          report.CompletedAt.Time,
		FailedAt:             report.FailedAt.Time,
		CreatedAt:            report.CreatedAt,
		ErrorMessage:         report.ErrorMessage.String,
	}

	jsonResponse(w, http.StatusCreated, reportResponse, "Report created successfully")
}

func (s *server) GetReportHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := UserFromContext(r)
	if !ok {
		errorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	reportIdStr := chi.URLParam(r, "reportId")
	reportId, err := uuid.Parse(reportIdStr)
	if err != nil {
		s.logger.Error("Error parsing report ID", err)
		errorResponse(w, http.StatusBadRequest, "Invalid report ID")
		return
	}

	report, err := s.store.GetReport(r.Context(), db.GetReportParams{
		ID:     reportId,
		UserID: user.ID,
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			errorResponse(w, http.StatusNotFound, "Report not found")
			return
		}
		s.logger.Error("Error getting report", err)
		errorResponse(w, http.StatusInternalServerError, "Error getting report")
		return
	}

	if report.CompletedAt.Valid {
		refreshNeeded := report.DownloadExpiresAt.Valid && report.DownloadExpiresAt.Time.Before(time.Now())
		if !report.DownloadUrl.Valid || refreshNeeded {
			expiredAt := time.Now().Add(time.Minute * 10)
			signedUrl, err := s.presignedClient.PresignGetObject(r.Context(), &s3.GetObjectInput{
				Bucket: aws.String(s.config.S3_BUCKET),
				Key:    aws.String(report.OutputFilePath.String),
			}, func(options *s3.PresignOptions) {
				options.Expires = time.Second * 10
			})

			s.logger.Debug("Generated presigned URL:", signedUrl.URL)

			if err != nil {
				s.logger.Error("Error generating presigned URL", err)
				errorResponse(w, http.StatusInternalServerError, "Error generating presigned URL")
				return
			}

			// update the report
			report, err = s.store.UpdateReport(r.Context(), db.UpdateReportParams{
				ID:                report.ID,
				UserID:            report.UserID,
				DownloadUrl:       sql.NullString{String: signedUrl.URL, Valid: true},
				DownloadExpiresAt: sql.NullTime{Time: expiredAt, Valid: true},
			})

			if err != nil {
				s.logger.Error("Error updating report with download URL", err)
				errorResponse(w, http.StatusInternalServerError, "Error updating report with download URL")
				return
			}
		}

	}

	reportResponse := ReportResponse{
		ID:                   report.ID,
		ReportType:           report.ReportType,
		OutputFilePath:       report.OutputFilePath.String,
		DownloadURL:          report.DownloadUrl.String,
		DownloadUrlExpiresAt: report.DownloadExpiresAt.Time,
		StartedAt:            report.StartedAt.Time,
		Status:               GetStatus(report),
		CompletedAt:          report.CompletedAt.Time,
		FailedAt:             report.FailedAt.Time,
		ErrorMessage:         report.ErrorMessage.String,
	}

	jsonResponse(w, http.StatusOK, reportResponse, "Report retrieved successfully")
}

func isDone(r db.Report) bool {
	return r.CompletedAt.Valid || r.FailedAt.Valid

}

func GetStatus(r db.Report) string {
	switch {
	case !r.StartedAt.Valid:
		return "requested"
	case r.StartedAt.Valid && !isDone(r):
		return "processing"
	case r.CompletedAt.Valid:
		return "completed"
	case r.FailedAt.Valid:
		return "failed"
	default:
		return "unknown"
	}
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

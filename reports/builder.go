package reports

import (
	"bytes"
	"compress/gzip"
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/trenchesdeveloper/csv-reporter/config"
	db "github.com/trenchesdeveloper/csv-reporter/db/sqlc"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"time"
)

type ReportBuilder struct {
	store     db.Store
	lozClient *LozClient
	s3Client  *s3.Client
	config    *config.AppConfig
	logger    *zap.SugaredLogger
}

func NewReportBuilder(store db.Store, lozClient *LozClient, s3Client *s3.Client, config *config.AppConfig, logger *zap.SugaredLogger) *ReportBuilder {
	return &ReportBuilder{
		store:     store,
		lozClient: lozClient,
		s3Client:  s3Client,
		config:    config,
		logger:    logger,
	}
}

func (rb *ReportBuilder) BuildReport(ctx context.Context, userId uuid.UUID, reportId uuid.UUID) (report db.Report, err error) {
	// Fetch the report from the database
	report, err = rb.store.GetReport(ctx, db.GetReportParams{
		ID:     reportId,
		UserID: userId,
	})
	if err != nil {
		return db.Report{}, fmt.Errorf("failed to get report %s: %w", reportId, err)
	}
	// Check if the report is already built
	if report.StartedAt.Valid {
		return report, nil
	}

	defer func() {
		if err != nil {
			// If an error occurs, update the report with the error message
			_, updateErr := rb.store.UpdateReport(ctx, db.UpdateReportParams{
				ID:           report.ID,
				UserID:       report.UserID,
				FailedAt:     sql.NullTime{Time: time.Now(), Valid: true},
				ErrorMessage: sql.NullString{String: err.Error(), Valid: true},
			})
			if updateErr != nil {
				err = fmt.Errorf("failed to update report with error: %w", updateErr)
			}
		}
	}()

	now := time.Now()
	// Update the report
	updatedReport, err := rb.store.UpdateReport(ctx, db.UpdateReportParams{
		ID:                report.ID,
		UserID:            report.UserID,
		StartedAt:         sql.NullTime{Time: now, Valid: true},
		CompletedAt:       sql.NullTime{Valid: false},
		FailedAt:          sql.NullTime{Valid: false},
		ErrorMessage:      sql.NullString{Valid: false},
		DownloadUrl:       sql.NullString{Valid: false},
		DownloadExpiresAt: sql.NullTime{Valid: false},
		OutputFilePath:    sql.NullString{Valid: false},
	})

	if err != nil {
		return db.Report{}, fmt.Errorf("failed to update report %s: %w", reportId, err)

	}

	resp, err := rb.lozClient.GetMonsters()

	if err != nil {
		return db.Report{}, fmt.Errorf("failed to fetch monsters: %w", err)
	}
	if len(resp.Data) == 0 {
		return db.Report{}, fmt.Errorf("no monsters found")
	}

	var buffer bytes.Buffer
	qzipWriter := gzip.NewWriter(&buffer)
	csvWriter := csv.NewWriter(qzipWriter)
	header := []string{
		"name",
		"id",
		"category",
		"description",
		"image",
		"common_locations",
		"drops",
		"dlc",
	}
	if err := csvWriter.Write(header); err != nil {
		return db.Report{}, fmt.Errorf("failed to write CSV header: %w", err)
	}
	for _, monster := range resp.Data {
		if err := csvWriter.Write([]string{
			monster.Name,
			fmt.Sprintf("%d", monster.Id),
			monster.Category,
			monster.Description,
			monster.Image,
			strings.Join(monster.Drops, ", "),
			strings.Join(monster.CommonLocations, ", "),
			strconv.FormatBool(monster.Dlc),
		}); err != nil {
			return db.Report{}, fmt.Errorf("failed to write CSV row: %w", err)
		}
	}
	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		return db.Report{}, fmt.Errorf("failed to flush CSV writer: %w", err)
	}

	if err := qzipWriter.Close(); err != nil {
		return db.Report{}, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	// prepare the file for S3 upload
	key := "/users/" + userId.String() + "/reports/" + reportId.String() + ".csv.gz"

	// Upload the file to S3
	_, err = rb.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(rb.config.S3_BUCKET),
		Key:    aws.String(key),
		Body:   bytes.NewReader(buffer.Bytes()),
	})
	if err != nil {
		return db.Report{}, fmt.Errorf("failed to upload report to S3: %w", err)
	}

	// Update the report with the download URL and expiration
	now = time.Now()
	updatedReport, err = rb.store.UpdateReport(ctx, db.UpdateReportParams{
		ID:             report.ID,
		UserID:         report.UserID,
		OutputFilePath: sql.NullString{String: key, Valid: true},
		CompletedAt:    sql.NullTime{Time: now, Valid: true},
	})

	if err != nil {
		return db.Report{}, fmt.Errorf("failed to update report %s: %w", reportId, err)
	}

	rb.logger.Info("successfully uploaded report to S3")

	return updatedReport, nil
}

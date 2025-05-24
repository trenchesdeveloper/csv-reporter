package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func createRandomReport(t *testing.T) Report {
	user := createRandomUser(t)
	
	arg := CreateReportParams{
		UserID:         user.ID,
		ReportType:     "csv",
		OutputFilePath: sql.NullString{String: "/path/to/report.csv", Valid: true},
		DownloadUrl:    sql.NullString{String: "https://example.com/report.csv", Valid: true},
		DownloadExpiresAt: sql.NullTime{
			Time:  time.Now().Add(24 * time.Hour),
			Valid: true,
		},
		ErrorMessage: sql.NullString{Valid: false},
		StartedAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
		FailedAt:    sql.NullTime{Valid: false},
		CompletedAt: sql.NullTime{Valid: false},
	}

	report, err := testStore.CreateReport(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, report)

	return report
}

func TestCreateReport(t *testing.T) {
	user := createRandomUser(t)
	
	arg := CreateReportParams{
		UserID:         user.ID,
		ReportType:     "csv",
		OutputFilePath: sql.NullString{String: "/path/to/report.csv", Valid: true},
		DownloadUrl:    sql.NullString{String: "https://example.com/report.csv", Valid: true},
		DownloadExpiresAt: sql.NullTime{
			Time:  time.Now().Add(24 * time.Hour),
			Valid: true,
		},
		ErrorMessage: sql.NullString{Valid: false},
		StartedAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
		FailedAt:    sql.NullTime{Valid: false},
		CompletedAt: sql.NullTime{Valid: false},
	}

	report, err := testStore.CreateReport(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, report)

	require.Equal(t, arg.UserID, report.UserID)
	require.Equal(t, arg.ReportType, report.ReportType)
	require.Equal(t, arg.OutputFilePath, report.OutputFilePath)
	require.Equal(t, arg.DownloadUrl, report.DownloadUrl)
	require.Equal(t, arg.DownloadExpiresAt.Time.Unix(), report.DownloadExpiresAt.Time.Unix())
	require.Equal(t, arg.ErrorMessage, report.ErrorMessage)
	require.Equal(t, arg.StartedAt.Time.Unix(), report.StartedAt.Time.Unix())
	require.Equal(t, arg.FailedAt, report.FailedAt)
	require.Equal(t, arg.CompletedAt, report.CompletedAt)

	require.NotZero(t, report.ID)
	require.NotZero(t, report.CreatedAt)
}

func TestGetReport(t *testing.T) {
	report1 := createRandomReport(t)
	
	arg := GetReportParams{
		UserID: report1.UserID,
		ID:     report1.ID,
	}

	report2, err := testStore.GetReport(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, report2)

	require.Equal(t, report1.ID, report2.ID)
	require.Equal(t, report1.UserID, report2.UserID)
	require.Equal(t, report1.ReportType, report2.ReportType)
	require.Equal(t, report1.OutputFilePath, report2.OutputFilePath)
	require.Equal(t, report1.DownloadUrl, report2.DownloadUrl)
	require.Equal(t, report1.DownloadExpiresAt.Time.Unix(), report2.DownloadExpiresAt.Time.Unix())
	require.Equal(t, report1.ErrorMessage, report2.ErrorMessage)
	require.Equal(t, report1.CreatedAt, report2.CreatedAt)
	require.Equal(t, report1.StartedAt.Time.Unix(), report2.StartedAt.Time.Unix())
	require.Equal(t, report1.FailedAt, report2.FailedAt)
	require.Equal(t, report1.CompletedAt, report2.CompletedAt)
}

func TestUpdateReport(t *testing.T) {
	report1 := createRandomReport(t)
	
	// Update with completed status
	arg := UpdateReportParams{
		UserID:         report1.UserID,
		ID:             report1.ID,
		OutputFilePath: report1.OutputFilePath,
		DownloadUrl:    report1.DownloadUrl,
		DownloadExpiresAt: report1.DownloadExpiresAt,
		ErrorMessage:   report1.ErrorMessage,
		StartedAt:      report1.StartedAt,
		FailedAt:       report1.FailedAt,
		CompletedAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
	}

	report2, err := testStore.UpdateReport(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, report2)

	require.Equal(t, report1.ID, report2.ID)
	require.Equal(t, report1.UserID, report2.UserID)
	require.Equal(t, report1.ReportType, report2.ReportType)
	require.Equal(t, report1.OutputFilePath, report2.OutputFilePath)
	require.Equal(t, report1.DownloadUrl, report2.DownloadUrl)
	require.Equal(t, report1.DownloadExpiresAt.Time.Unix(), report2.DownloadExpiresAt.Time.Unix())
	require.Equal(t, report1.ErrorMessage, report2.ErrorMessage)
	require.Equal(t, report1.CreatedAt, report2.CreatedAt)
	require.Equal(t, report1.StartedAt.Time.Unix(), report2.StartedAt.Time.Unix())
	require.Equal(t, report1.FailedAt, report2.FailedAt)
	require.True(t, report2.CompletedAt.Valid)
	require.NotZero(t, report2.CompletedAt.Time)
}

func TestDeleteReport(t *testing.T) {
	report1 := createRandomReport(t)
	
	arg := DeleteReportParams{
		UserID: report1.UserID,
		ID:     report1.ID,
	}

	err := testStore.DeleteReport(context.Background(), arg)
	require.NoError(t, err)

	// Verify it was deleted
	getArg := GetReportParams{
		UserID: report1.UserID,
		ID:     report1.ID,
	}
	report2, err := testStore.GetReport(context.Background(), getArg)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, report2)
}

func TestNonExistentReport(t *testing.T) {
	// Try to get a report with a non-existent ID
	randomUserID := uuid.New()
	randomReportID := uuid.New()
	
	arg := GetReportParams{
		UserID: randomUserID,
		ID:     randomReportID,
	}

	report, err := testStore.GetReport(context.Background(), arg)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, report)
}

package reports

import "github.com/google/uuid"


type SQSMessage struct {
	ReportID uuid.UUID `json:"report_id"`
	UserID   uuid.UUID `json:"user_id"`
}
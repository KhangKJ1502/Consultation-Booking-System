package entity

import (
	"time"

	"github.com/google/uuid"
)

type BackgroundJob struct {
	JobID        uuid.UUID  `json:"job_id" db:"job_id"`
	JobType      string     `json:"job_type" db:"job_type"`
	JobPayload   []byte     `json:"job_payload" db:"job_payload"`
	JobStatus    string     `json:"job_status" db:"job_status"`
	AttemptCount int        `json:"attempt_count" db:"attempt_count"`
	MaxAttempts  int        `json:"max_attempts" db:"max_attempts"`
	ScheduledAt  time.Time  `json:"scheduled_at" db:"scheduled_at"`
	StartedAt    *time.Time `json:"started_at,omitempty" db:"started_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty" db:"completed_at"`
	ErrorMessage *string    `json:"error_message,omitempty" db:"error_message"`
	JobCreatedAt time.Time  `json:"job_created_at" db:"job_created_at"`
}

// TableName returns the table name for GORM
func (BackgroundJob) TableName() string {
	return "tbl_background_jobs"
}

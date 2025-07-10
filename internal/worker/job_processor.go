package worker

import (
	entityBackground "cbs_backend/internal/modules/background_job/entity"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// JobProcessor handles recording job execution lifecycle in database
type JobProcessor struct {
	db *gorm.DB
}

// NewJobProcessor creates a new instance of JobProcessor
func NewJobProcessor(db *gorm.DB) *JobProcessor {
	return &JobProcessor{db: db}
}

// RecordJobStart records the start of a job execution
func (jp *JobProcessor) RecordJobStart(job Job) error {
	now := time.Now()

	// Safely serialize payload to JSON
	payloadBytes := jp.serializePayload(job.Payload)

	// Create background job record
	bgJob := entityBackground.BackgroundJob{
		JobID:        uuid.New(),
		JobType:      job.Type,
		JobPayload:   payloadBytes,
		JobStatus:    "processing",
		AttemptCount: job.RetryCount,
		MaxAttempts:  job.MaxRetries,
		ScheduledAt:  job.CreatedAt,
		StartedAt:    &now,
	}

	return jp.db.Create(&bgJob).Error
}

// RecordJobSuccess marks a job as successfully completed
func (jp *JobProcessor) RecordJobSuccess(job Job) error {
	now := time.Now()

	updates := map[string]interface{}{
		"job_status":   "completed",
		"completed_at": &now,
	}

	return jp.updateJobByTypeAndStatus(job.Type, "processing", updates)
}

// RecordJobFailure marks a job as failed with error details
func (jp *JobProcessor) RecordJobFailure(job Job, err error) error {
	updates := map[string]interface{}{
		"job_status":    "failed",
		"error_message": err.Error(),
		"attempt_count": job.RetryCount,
	}

	return jp.updateJobByTypeAndStatus(job.Type, "processing", updates)
}

// serializePayload safely converts payload to JSON bytes
func (jp *JobProcessor) serializePayload(payload interface{}) []byte {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		// Return empty JSON object if serialization fails
		return []byte("{}")
	}
	return payloadBytes
}

// updateJobByTypeAndStatus updates job record by type and current status
func (jp *JobProcessor) updateJobByTypeAndStatus(jobType, currentStatus string, updates map[string]interface{}) error {
	return jp.db.Model(&entityBackground.BackgroundJob{}).
		Where("job_type = ? AND job_status = ?", jobType, currentStatus).
		Updates(updates).Error
}

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
func (jp *JobProcessor) RecordJobResult(job Job, err error) error {
	if err != nil {
		return jp.RecordJobFailure(job, err)
	}
	return jp.RecordJobSuccess(job)
}

// package worker

// import (
// 	entityBackground "cbs_backend/internal/modules/background_job/entity"
// 	"encoding/json"
// 	"fmt"
// 	"time"

// 	"github.com/google/uuid"
// 	"gorm.io/gorm"
// )

// // =====================================================================
// // JOB PROCESSOR STRUCT
// // =====================================================================

// type JobProcessor struct {
// 	db *gorm.DB
// }

// // =====================================================================
// // CONSTRUCTOR
// // =====================================================================

// func NewJobProcessor(db *gorm.DB) *JobProcessor {
// 	return &JobProcessor{db: db}
// }

// // =====================================================================
// // UNIFIED JOB RESULT RECORDING
// // =====================================================================

// // RecordJobResult - Phương thức thống nhất để ghi lại kết quả job
// func (jp *JobProcessor) RecordJobResult(job Job, err error) error {
// 	now := time.Now()

// 	if err != nil {
// 		return jp.recordJobFailure(job, err, now)
// 	}
// 	return jp.recordJobSuccess(job, now)
// }

// // RecordJobStart - Ghi lại việc bắt đầu thực thi job
// func (jp *JobProcessor) RecordJobStart(job Job) error {
// 	now := time.Now()

// 	// Serialize payload safely
// 	payloadBytes := jp.serializePayload(job.Payload)

// 	// Create background job record
// 	bgJob := entityBackground.BackgroundJob{
// 		JobID:        uuid.New(),
// 		JobType:      job.Type,
// 		JobPayload:   payloadBytes,
// 		JobStatus:    "processing",
// 		AttemptCount: job.RetryCount,
// 		MaxAttempts:  job.MaxRetries,
// 		ScheduledAt:  job.CreatedAt,
// 		StartedAt:    &now,
// 	}

// 	return jp.db.Create(&bgJob).Error
// }

// // =====================================================================
// // PRIVATE HELPER METHODS
// // =====================================================================

// // recordJobSuccess - Ghi lại job thành công
// func (jp *JobProcessor) recordJobSuccess(job Job, completedAt time.Time) error {
// 	updates := map[string]interface{}{
// 		"job_status":   "completed",
// 		"completed_at": &completedAt,
// 	}

// 	conditions := map[string]interface{}{
// 		"job_type":   job.Type,
// 		"job_status": "processing",
// 	}

// 	return jp.updateJobByConditions(conditions, updates)
// }

// // recordJobFailure - Ghi lại job thất bại
// func (jp *JobProcessor) recordJobFailure(job Job, err error, failedAt time.Time) error {
// 	updates := map[string]interface{}{
// 		"job_status":    "failed",
// 		"error_message": err.Error(),
// 		"attempt_count": job.RetryCount,
// 		"failed_at":     &failedAt,
// 	}

// 	conditions := map[string]interface{}{
// 		"job_type":   job.Type,
// 		"job_status": "processing",
// 	}

// 	return jp.updateJobByConditions(conditions, updates)
// }

// // =====================================================================
// // GENERIC UPDATE METHODS
// // =====================================================================

// // updateJobByConditions - Phương thức cập nhật chung với nhiều điều kiện
// func (jp *JobProcessor) updateJobByConditions(conditions map[string]interface{}, updates map[string]interface{}) error {
// 	query := jp.db.Model(&entityBackground.BackgroundJob{})

// 	// Apply all conditions
// 	for key, value := range conditions {
// 		query = query.Where(key+" = ?", value)
// 	}

// 	result := query.Updates(updates)
// 	if result.Error != nil {
// 		return fmt.Errorf("không thể cập nhật job: %w", result.Error)
// 	}

// 	if result.RowsAffected == 0 {
// 		return fmt.Errorf("không tìm thấy job để cập nhật với điều kiện: %v", conditions)
// 	}

// 	return nil
// }

// // updateJobByID - Cập nhật job bằng ID
// func (jp *JobProcessor) updateJobByID(jobID string, updates map[string]interface{}) error {
// 	conditions := map[string]interface{}{
// 		"job_id": jobID,
// 	}
// 	return jp.updateJobByConditions(conditions, updates)
// }

// // updateJobByType - Cập nhật job bằng type
// func (jp *JobProcessor) updateJobByType(jobType string, updates map[string]interface{}) error {
// 	conditions := map[string]interface{}{
// 		"job_type": jobType,
// 	}
// 	return jp.updateJobByConditions(conditions, updates)
// }

// // =====================================================================
// // UTILITY METHODS
// // =====================================================================

// // serializePayload - Chuyển đổi payload thành JSON bytes một cách an toàn
// func (jp *JobProcessor) serializePayload(payload interface{}) []byte {
// 	if payload == nil {
// 		return []byte("{}")
// 	}

// 	payloadBytes, err := json.Marshal(payload)
// 	if err != nil {
// 		// Trả về empty JSON object nếu serialization thất bại
// 		return []byte("{}")
// 	}
// 	return payloadBytes
// }

// // =====================================================================
// // QUERY METHODS
// // =====================================================================

// // GetJobByID - Lấy job theo ID
// func (jp *JobProcessor) GetJobByID(jobID string) (*entityBackground.BackgroundJob, error) {
// 	var job entityBackground.BackgroundJob
// 	err := jp.db.Where("job_id = ?", jobID).First(&job).Error
// 	if err != nil {
// 		return nil, fmt.Errorf("không tìm thấy job với ID %s: %w", jobID, err)
// 	}
// 	return &job, nil
// }

// // GetJobsByType - Lấy danh sách jobs theo type
// func (jp *JobProcessor) GetJobsByType(jobType string) ([]entityBackground.BackgroundJob, error) {
// 	var jobs []entityBackground.BackgroundJob
// 	err := jp.db.Where("job_type = ?", jobType).Find(&jobs).Error
// 	if err != nil {
// 		return nil, fmt.Errorf("không thể lấy jobs với type %s: %w", jobType, err)
// 	}
// 	return jobs, nil
// }

// // GetJobsByStatus - Lấy danh sách jobs theo status
// func (jp *JobProcessor) GetJobsByStatus(status string) ([]entityBackground.BackgroundJob, error) {
// 	var jobs []entityBackground.BackgroundJob
// 	err := jp.db.Where("job_status = ?", status).Find(&jobs).Error
// 	if err != nil {
// 		return nil, fmt.Errorf("không thể lấy jobs với status %s: %w", status, err)
// 	}
// 	return jobs, nil
// }

// // GetFailedJobs - Lấy danh sách jobs thất bại
// func (jp *JobProcessor) GetFailedJobs() ([]entityBackground.BackgroundJob, error) {
// 	return jp.GetJobsByStatus("failed")
// }

// // GetPendingJobs - Lấy danh sách jobs đang chờ
// func (jp *JobProcessor) GetPendingJobs() ([]entityBackground.BackgroundJob, error) {
// 	return jp.GetJobsByStatus("pending")
// }

// // GetProcessingJobs - Lấy danh sách jobs đang xử lý
// func (jp *JobProcessor) GetProcessingJobs() ([]entityBackground.BackgroundJob, error) {
// 	return jp.GetJobsByStatus("processing")
// }

// // =====================================================================
// // BULK OPERATIONS
// // =====================================================================

// // BulkUpdateJobStatus - Cập nhật trạng thái hàng loạt jobs
// func (jp *JobProcessor) BulkUpdateJobStatus(jobIDs []string, newStatus string) error {
// 	if len(jobIDs) == 0 {
// 		return nil
// 	}

// 	updates := map[string]interface{}{
// 		"job_status": newStatus,
// 		"updated_at": time.Now(),
// 	}

// 	result := jp.db.Model(&entityBackground.BackgroundJob{}).
// 		Where("job_id IN ?", jobIDs).
// 		Updates(updates)

// 	if result.Error != nil {
// 		return fmt.Errorf("không thể cập nhật bulk jobs: %w", result.Error)
// 	}

// 	return nil
// }

// // RetryFailedJobs - Thử lại các jobs thất bại
// func (jp *JobProcessor) RetryFailedJobs(maxAge time.Duration) error {
// 	cutoffTime := time.Now().Add(-maxAge)

// 	updates := map[string]interface{}{
// 		"job_status":    "pending",
// 		"error_message": nil,
// 		"updated_at":    time.Now(),
// 	}

// 	result := jp.db.Model(&entityBackground.BackgroundJob{}).
// 		Where("job_status = ? AND failed_at < ?", "failed", cutoffTime).
// 		Updates(updates)

// 	if result.Error != nil {
// 		return fmt.Errorf("không thể retry failed jobs: %w", result.Error)
// 	}

// 	return nil
// }

// // CleanupOldJobs - Dọn dẹp các jobs cũ
// func (jp *JobProcessor) CleanupOldJobs(maxAge time.Duration) error {
// 	cutoffTime := time.Now().Add(-maxAge)

// 	result := jp.db.Where("completed_at < ? OR (failed_at < ? AND job_status = 'failed')",
// 		cutoffTime, cutoffTime).
// 		Delete(&entityBackground.BackgroundJob{})

// 	if result.Error != nil {
// 		return fmt.Errorf("không thể cleanup old jobs: %w", result.Error)
// 	}

// 	return nil
// }

// // =====================================================================
// // STATISTICS AND MONITORING
// // =====================================================================

// // GetJobStats - Lấy thống kê jobs
// func (jp *JobProcessor) GetJobStats() (map[string]interface{}, error) {
// 	stats := make(map[string]interface{})

// 	// Count by status
// 	statusCounts := make(map[string]int64)
// 	rows, err := jp.db.Model(&entityBackground.BackgroundJob{}).
// 		Select("job_status, COUNT(*) as count").
// 		Group("job_status").
// 		Rows()

// 	if err != nil {
// 		return nil, fmt.Errorf("không thể lấy job stats: %w", err)
// 	}
// 	defer rows.Close()

// 	for rows.Next() {
// 		var status string
// 		var count int64
// 		if err := rows.Scan(&status, &count); err != nil {
// 			return nil, err
// 		}
// 		statusCounts[status] = count
// 	}

// 	stats["status_counts"] = statusCounts

// 	// Count by type
// 	typeCounts := make(map[string]int64)
// 	rows, err = jp.db.Model(&entityBackground.BackgroundJob{}).
// 		Select("job_type, COUNT(*) as count").
// 		Group("job_type").
// 		Rows()

// 	if err != nil {
// 		return nil, fmt.Errorf("không thể lấy job type stats: %w", err)
// 	}
// 	defer rows.Close()

// 	for rows.Next() {
// 		var jobType string
// 		var count int64
// 		if err := rows.Scan(&jobType, &count); err != nil {
// 			return nil, err
// 		}
// 		typeCounts[jobType] = count
// 	}

// 	stats["type_counts"] = typeCounts

// 	// Total jobs
// 	var totalJobs int64
// 	if err := jp.db.Model(&entityBackground.BackgroundJob{}).Count(&totalJobs).Error; err != nil {
// 		return nil, fmt.Errorf("không thể đếm total jobs: %w", err)
// 	}
// 	stats["total_jobs"] = totalJobs

// 	return stats, nil
// }

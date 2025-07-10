package worker

import (
	"cbs_backend/internal/service/interfaces"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

// =====================================================================
// CONSTANTS AND CONFIGURATION
// =====================================================================

const (
	DefaultQueueSize          = 1000
	NotificationCheckInterval = 5 * time.Second
	MaxQueueFullness          = 2 // maxWorkers * 2
	TestStartupDelay          = 5 * time.Second
)

// =====================================================================
// MAIN STRUCT DEFINITION
// =====================================================================

type WorkerScheduler struct {
	// Core components
	db     *gorm.DB
	cron   *cron.Cron
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Job processing
	jobQueue   chan Job
	maxWorkers int
	activeJobs int
	jobMutex   sync.RWMutex

	// Services
	reminderSvc       *ReminderService
	cleanupSvc        *CleanupService
	notificationSvc   *NotificationService
	enhancedNotifySvc *EnhancedNotificationService
	jobProcessor      *JobProcessor
}

// =====================================================================
// CONSTRUCTOR
// =====================================================================

func NewWorkerScheduler(db *gorm.DB, maxWorkers int, emailService interfaces.EmailService, redisClient *redis.Client) *WorkerScheduler {
	ctx, cancel := context.WithCancel(context.Background())

	// Configure cron parser
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	cronScheduler := cron.New(
		cron.WithParser(parser),
		cron.WithChain(cron.Recover(cron.DefaultLogger)),
	)

	ws := &WorkerScheduler{
		// Core components
		db:         db,
		cron:       cronScheduler,
		ctx:        ctx,
		cancel:     cancel,
		jobQueue:   make(chan Job, DefaultQueueSize),
		maxWorkers: maxWorkers,
	}

	// Initialize services
	ws.initializeServices(db, emailService, redisClient)

	return ws
}

func (ws *WorkerScheduler) initializeServices(db *gorm.DB, emailService interfaces.EmailService, redisClient *redis.Client) {
	ws.reminderSvc = NewReminderService(db, emailService)
	ws.cleanupSvc = NewCleanupService(db)
	ws.notificationSvc = NewNotificationService(db)
	ws.enhancedNotifySvc = NewEnhancedNotificationService(db, redisClient, emailService)
	ws.jobProcessor = NewJobProcessor(db)
}

// =====================================================================
// LIFECYCLE MANAGEMENT
// =====================================================================

func (ws *WorkerScheduler) Start() error {
	log.Println("🚀 Starting Worker Scheduler...")

	// Start worker goroutines
	if err := ws.startWorkers(); err != nil {
		return err
	}

	// Start notification processor
	go ws.notificationProcessor()

	// Schedule all cron jobs
	if err := ws.scheduleCronJobs(); err != nil {
		return fmt.Errorf("failed to schedule cron jobs: %w", err)
	}

	// Start cron scheduler
	ws.cron.Start()

	// Log startup information
	ws.logStartupInfo()

	log.Printf("✅ Worker Scheduler started with %d workers", ws.maxWorkers)
	return nil
}

func (ws *WorkerScheduler) Stop() {
	log.Println("🛑 Stopping Worker Scheduler...")

	// Stop cron scheduler and wait for completion
	ctx := ws.cron.Stop()
	<-ctx.Done()

	// Signal all goroutines to stop
	ws.cancel()

	// Close job queue to signal workers
	close(ws.jobQueue)

	// Wait for all workers to finish
	ws.wg.Wait()

	log.Println("✅ Worker Scheduler stopped")
}

// =====================================================================
// WORKER MANAGEMENT
// =====================================================================

func (ws *WorkerScheduler) startWorkers() error {
	for i := 0; i < ws.maxWorkers; i++ {
		ws.wg.Add(1)
		go ws.worker(i)
	}
	return nil
}

func (ws *WorkerScheduler) worker(id int) {
	defer ws.wg.Done()
	log.Printf("👷 Worker %d started", id)

	for {
		select {
		case job, ok := <-ws.jobQueue:
			if !ok {
				log.Printf("👷 Worker %d stopped - queue closed", id)
				return
			}
			ws.handleJob(job, id)

		case <-ws.ctx.Done():
			log.Printf("👷 Worker %d stopped by context", id)
			return
		}
	}
}

func (ws *WorkerScheduler) handleJob(job Job, workerID int) {
	// Update active job count
	ws.updateActiveJobs(1)
	defer ws.updateActiveJobs(-1)

	// Process the job
	ws.processJob(job, workerID)
}

func (ws *WorkerScheduler) updateActiveJobs(delta int) {
	ws.jobMutex.Lock()
	ws.activeJobs += delta
	ws.jobMutex.Unlock()
}

// =====================================================================
// JOB PROCESSING
// =====================================================================

func (ws *WorkerScheduler) processJob(job Job, workerID int) {
	log.Printf("⚙️ Worker %d processing job: %s (type: %s)", workerID, job.ID, job.Type)

	// Record job start
	if err := ws.jobProcessor.RecordJobStart(job); err != nil {
		log.Printf("⚠️ Failed to record job start: %v", err)
	}

	// Execute job based on type
	err := ws.executeJob(job)

	// Handle job result
	ws.handleJobResult(job, err)
}

func (ws *WorkerScheduler) executeJob(job Job) error {
	switch job.Type {
	case "booking_reminder":
		log.Println("Bắt đầu gọi hàm sender mind//////////////")
		return ws.reminderSvc.SendBookingReminders()

	case "check_missed_bookings":
		return ws.reminderSvc.CheckMissedBookings()

	case "handle_duplicate_bookings":
		return ws.reminderSvc.HandleDuplicateBookings()

	case "cleanup_old_data":
		days := ws.extractCleanupDays(job.Payload)
		return ws.cleanupSvc.CleanupOldData(days)

	case "weekly_statistics":
		return ws.reminderSvc.GenerateWeeklyStatistics()

	case "send_email_batch":
		return ws.notificationSvc.ProcessEmailBatch(job.Payload)

	case "process_notifications":
		return ws.processNotifications()

	case "send_email", "send_telegram", "send_sms":
		return ws.enhancedNotifySvc.ProcessNotificationJob(job)

	default:
		return fmt.Errorf("unknown job type: %s", job.Type)
	}
}

func (ws *WorkerScheduler) extractCleanupDays(payload interface{}) int {
	defaultDays := 30
	payloadMap, ok := payload.(map[string]interface{})
	if !ok {
		return defaultDays
	}

	if days, exists := payloadMap["days"]; exists {
		if daysInt, ok := days.(int); ok {
			return daysInt
		}
	}
	return defaultDays
}

func (ws *WorkerScheduler) handleJobResult(job Job, err error) {
	if err != nil {
		ws.handleJobFailure(job, err)
	} else {
		ws.handleJobSuccess(job)
	}
}

func (ws *WorkerScheduler) handleJobFailure(job Job, err error) {
	log.Printf("❌ Job %s failed: %v", job.ID, err)

	// Record failure
	if recordErr := ws.jobProcessor.RecordJobFailure(job, err); recordErr != nil {
		log.Printf("⚠️ Failed to record job failure: %v", recordErr)
	}

	// Retry if possible
	if job.RetryCount < job.MaxRetries {
		ws.scheduleRetry(job)
	} else {
		log.Printf("💀 Job %s failed permanently after %d attempts", job.ID, job.MaxRetries)
	}
}

func (ws *WorkerScheduler) handleJobSuccess(job Job) {
	log.Printf("✅ Job %s completed successfully", job.ID)

	if recordErr := ws.jobProcessor.RecordJobSuccess(job); recordErr != nil {
		log.Printf("⚠️ Failed to record job success: %v", recordErr)
	}
}

func (ws *WorkerScheduler) scheduleRetry(job Job) {
	job.RetryCount++
	retryDelay := time.Duration(job.RetryCount*job.RetryCount) * time.Minute

	log.Printf("🔄 Retrying job %s in %v (attempt %d/%d)",
		job.ID, retryDelay, job.RetryCount, job.MaxRetries)

	go func() {
		time.Sleep(retryDelay)
		ws.AddJob(job)
	}()
}

// =====================================================================
// JOB QUEUE MANAGEMENT
// =====================================================================

func (ws *WorkerScheduler) AddJob(job Job) {
	// Check if scheduler is shutting down
	select {
	case <-ws.ctx.Done():
		log.Printf("⚠️ Cannot add job %s: scheduler is shutting down", job.ID)
		return
	default:
	}

	// Check queue capacity
	if ws.isQueueFull() {
		log.Printf("⚠️ Queue is full, dropping job: %s", job.ID)
		return
	}

	// Add job to queue
	select {
	case ws.jobQueue <- job:
		log.Printf("📝 Job added to queue: %s (type: %s)", job.ID, job.Type)
	case <-ws.ctx.Done():
		log.Printf("⚠️ Cannot add job %s: scheduler is shutting down", job.ID)
	default:
		log.Printf("⚠️ Queue is full, dropping job: %s", job.ID)
	}
}

func (ws *WorkerScheduler) isQueueFull() bool {
	ws.jobMutex.RLock()
	defer ws.jobMutex.RUnlock()
	return ws.activeJobs >= ws.maxWorkers*MaxQueueFullness
}

// =====================================================================
// CRON JOB SCHEDULING
// =====================================================================

func (ws *WorkerScheduler) scheduleCronJobs() error {
	log.Println("📅 Scheduling cron jobs...")

	cronJobs := []struct {
		name     string
		schedule string
		jobType  string
		payload  interface{}
		priority int
		retries  int
	}{
		{"process_notifications", "* * * * *", "process_notifications", nil, 1, 3},
		{"booking_reminder", "*/2 * * * *", "booking_reminder", nil, 1, 3},
		{"check_missed_bookings", "*/15 * * * *", "check_missed_bookings", nil, 2, 3},
		{"handle_duplicate_bookings", "*/30 * * * *", "handle_duplicate_bookings", nil, 2, 3},
		{"cleanup_old_data", "0 2 * * *", "cleanup_old_data", map[string]interface{}{"days": 30}, 3, 2},
		{"weekly_statistics", "0 6 * * 0", "weekly_statistics", nil, 2, 3},
	}

	for _, cronJob := range cronJobs {
		if err := ws.scheduleCronJob(cronJob.name, cronJob.schedule, cronJob.jobType, cronJob.payload, cronJob.priority, cronJob.retries); err != nil {
			return err
		}
	}

	// Schedule test job
	ws.scheduleTestJob()

	log.Println("📅 All cron jobs scheduled successfully")
	return nil
}

func (ws *WorkerScheduler) scheduleCronJob(name, schedule, jobType string, payload interface{}, priority, retries int) error {
	entryID, err := ws.cron.AddFunc(schedule, func() {
		log.Printf("⏰ Cron triggered: %s", name)
		ws.AddJob(Job{
			ID:         generateJobID(),
			Type:       jobType,
			Payload:    payload,
			Priority:   priority,
			MaxRetries: retries,
			CreatedAt:  time.Now(),
		})
	})

	if err != nil {
		return fmt.Errorf("failed to add %s job: %w", name, err)
	}

	log.Printf("✅ Scheduled %s (ID: %d) - %s", name, entryID, schedule)
	return nil
}

func (ws *WorkerScheduler) scheduleTestJob() {
	go func() {
		time.Sleep(TestStartupDelay)
		log.Println("⏰ Manual trigger: booking_reminder (test)")
		ws.AddJob(Job{
			ID:         generateJobID(),
			Type:       "booking_reminder",
			Payload:    nil,
			Priority:   1,
			MaxRetries: 3,
			CreatedAt:  time.Now(),
		})
	}()
}

// =====================================================================
// NOTIFICATION PROCESSING
// =====================================================================

func (ws *WorkerScheduler) notificationProcessor() {
	ticker := time.NewTicker(NotificationCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := ws.processNotifications(); err != nil {
				log.Printf("⚠️ Failed to process notifications: %v", err)
			}
		case <-ws.ctx.Done():
			log.Println("🛑 Notification processor stopped")
			return
		}
	}
}

func (ws *WorkerScheduler) processNotifications() error {
	notifications, err := ws.fetchPendingNotifications()
	if err != nil {
		return fmt.Errorf("failed to get pending notifications: %w", err)
	}

	if len(notifications) > 0 {
		log.Printf("📧 Found %d pending email notifications", len(notifications))
	}

	for _, notification := range notifications {
		if err := ws.processNotification(notification); err != nil {
			log.Printf("⚠️ Failed to process notification %s: %v", notification.NotificationID, err)
		}
	}

	return nil
}

func (ws *WorkerScheduler) fetchPendingNotifications() ([]struct {
	NotificationID string `json:"notification_id"`
	UserID         string `json:"user_id"`
	UserEmail      string `json:"user_email"`
	Title          string `json:"title"`
	Message        string `json:"message"`
	Data           string `json:"data"`
}, error) {
	var notifications []struct {
		NotificationID string `json:"notification_id"`
		UserID         string `json:"user_id"`
		UserEmail      string `json:"user_email"`
		Title          string `json:"title"`
		Message        string `json:"message"`
		Data           string `json:"data"`
	}

	query := `
        SELECT 
            sn.notification_id,
            sn.recipient_user_id as user_id,
            u.user_email,
            sn.notification_title as title,
            sn.notification_message as message,
            sn.notification_data as data
        FROM tbl_system_notifications sn
        JOIN tbl_users u ON sn.recipient_user_id = u.user_id
        WHERE sn.notification_status = 'pending'
        AND 'email' = ANY(sn.delivery_methods)
        AND sn.notification_created_at >= NOW() - INTERVAL '1 hour'
        ORDER BY sn.notification_created_at ASC
        LIMIT 50
    `

	err := ws.db.Raw(query).Scan(&notifications).Error
	return notifications, err
}

func (ws *WorkerScheduler) processNotification(notification struct {
	NotificationID string `json:"notification_id"`
	UserID         string `json:"user_id"`
	UserEmail      string `json:"user_email"`
	Title          string `json:"title"`
	Message        string `json:"message"`
	Data           string `json:"data"`
}) error {
	// Parse notification data
	dataMap, err := ws.parseNotificationData(notification.Data)
	if err != nil {
		return err
	}

	// Add essential fields
	dataMap["notification_id"] = notification.NotificationID
	dataMap["user_id"] = notification.UserID

	// Update status to processing
	if err := ws.updateNotificationStatus(notification.NotificationID, "processing"); err != nil {
		return err
	}

	// Create and add email job
	emailJob := ws.createEmailJob(notification, dataMap)
	ws.AddJob(emailJob)

	log.Printf("➕ Added email job for notification %s", notification.NotificationID)
	return nil
}

func (ws *WorkerScheduler) parseNotificationData(data string) (map[string]interface{}, error) {
	dataMap := make(map[string]interface{})

	if data != "" {
		if err := json.Unmarshal([]byte(data), &dataMap); err != nil {
			return nil, fmt.Errorf("failed to parse notification data: %w", err)
		}
	}

	return dataMap, nil
}

func (ws *WorkerScheduler) updateNotificationStatus(notificationID, status string) error {
	return ws.db.Exec(
		"UPDATE tbl_system_notifications SET notification_status = ? WHERE notification_id = ?",
		status, notificationID,
	).Error
}

func (ws *WorkerScheduler) createEmailJob(notification struct {
	NotificationID string `json:"notification_id"`
	UserID         string `json:"user_id"`
	UserEmail      string `json:"user_email"`
	Title          string `json:"title"`
	Message        string `json:"message"`
	Data           string `json:"data"`
}, dataMap map[string]interface{}) Job {
	return Job{
		ID:   generateJobID(),
		Type: "send_email",
		Payload: map[string]interface{}{
			"notification_id": notification.NotificationID,
			"user_id":         notification.UserID,
			"recipient":       notification.UserEmail,
			"subject":         notification.Title,
			"body":            notification.Message,
			"data":            dataMap,
			"template":        "notification",
		},
		Priority:   1,
		MaxRetries: 3,
		CreatedAt:  time.Now(),
	}
}

// =====================================================================
// UTILITY FUNCTIONS
// =====================================================================

func (ws *WorkerScheduler) logStartupInfo() {
	entries := ws.cron.Entries()
	log.Printf("📋 Total cron entries: %d", len(entries))
	for i, entry := range entries {
		log.Printf("  Entry %d: Next run at %v", i+1, entry.Next)
	}
}

func (ws *WorkerScheduler) GetStats() map[string]interface{} {
	ws.jobMutex.RLock()
	defer ws.jobMutex.RUnlock()

	return map[string]interface{}{
		"active_jobs":  ws.activeJobs,
		"max_workers":  ws.maxWorkers,
		"queue_length": len(ws.jobQueue),
		"queue_cap":    cap(ws.jobQueue),
	}
}

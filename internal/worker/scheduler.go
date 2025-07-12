package worker

import (
	"cbs_backend/internal/service/interfaces"
	"context"
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
	DefaultQueueSize = 1000
	MaxQueueFullness = 2
	TestStartupDelay = 5 * time.Second
	ShutdownTimeout  = 30 * time.Second
	RetryMultiplier  = 2
	BaseRetryDelay   = 30 * time.Second
	MaxRetryDelay    = 5 * time.Minute
)

// =====================================================================
// CONFIGURATION STRUCTURES
// =====================================================================

type WorkerConfig struct {
	MaxWorkers      int
	QueueSize       int
	ShutdownTimeout time.Duration
}

type CronJobConfig struct {
	Name     string
	Schedule string
	JobType  string
	Payload  interface{}
	Priority int
	Retries  int
}

// =====================================================================
// JOB PROCESSOR INTERFACE
// =====================================================================

type JobProcessorInterface interface {
	RecordJobStart(job Job) error
	RecordJobResult(job Job, err error) error
}

// =====================================================================
// JOB EXECUTOR INTERFACE
// =====================================================================

type JobExecutor interface {
	Execute(job Job) error
}

// =====================================================================
// JOB LOGGER
// =====================================================================

type JobLogger struct {
	workerID int
}

func NewJobLogger(workerID int) *JobLogger {
	return &JobLogger{workerID: workerID}
}

func (jl *JobLogger) LogStart(job Job) {
	log.Printf("⚙️ Worker %d processing job: %s (type: %s)", jl.workerID, job.ID, job.Type)
}

func (jl *JobLogger) LogSuccess(job Job) {
	log.Printf("✅ Job %s completed successfully", job.ID)
}

func (jl *JobLogger) LogFailure(job Job, err error) {
	log.Printf("❌ Job %s failed: %v", job.ID, err)
}

func (jl *JobLogger) LogRetry(job Job, retryDelay time.Duration) {
	log.Printf("🔄 Retrying job %s in %v (attempt %d/%d)",
		job.ID, retryDelay, job.RetryCount, job.MaxRetries)
}

func (jl *JobLogger) LogPermanentFailure(job Job) {
	log.Printf("💀 Job %s permanently failed after %d attempts", job.ID, job.MaxRetries)
}

// =====================================================================
// JOB RESULT PROCESSOR
// =====================================================================

type JobResultProcessor struct {
	processor JobProcessorInterface
	scheduler *WorkerScheduler
}

func NewJobResultProcessor(processor JobProcessorInterface, scheduler *WorkerScheduler) *JobResultProcessor {
	return &JobResultProcessor{
		processor: processor,
		scheduler: scheduler,
	}
}

func (jrp *JobResultProcessor) ProcessResult(job Job, err error, logger *JobLogger) {
	// Record result in database
	if recordErr := jrp.processor.RecordJobResult(job, err); recordErr != nil {
		log.Printf("⚠️ Failed to record job result: %v", recordErr)
	}

	if err != nil {
		jrp.handleJobError(job, err, logger)
	} else {
		jrp.handleJobSuccess(job, logger)
	}
}

func (jrp *JobResultProcessor) handleJobSuccess(job Job, logger *JobLogger) {
	logger.LogSuccess(job)
}

func (jrp *JobResultProcessor) handleJobError(job Job, err error, logger *JobLogger) {
	logger.LogFailure(job, err)

	if job.RetryCount < job.MaxRetries {
		jrp.scheduleRetry(job, logger)
	} else {
		logger.LogPermanentFailure(job)
	}
}

func (jrp *JobResultProcessor) scheduleRetry(job Job, logger *JobLogger) {
	job.RetryCount++

	// Calculate exponential backoff with jitter
	baseDelay := BaseRetryDelay * time.Duration(job.RetryCount)
	if baseDelay > MaxRetryDelay {
		baseDelay = MaxRetryDelay
	}

	// Add some jitter (±10%)
	jitter := time.Duration(float64(baseDelay) * 0.1 * float64(2*int64(time.Now().UnixNano()%2)-1))
	retryDelay := baseDelay + jitter

	logger.LogRetry(job, retryDelay)

	go func() {
		time.Sleep(retryDelay)
		jrp.scheduler.AddJob(job)
	}()
}

// =====================================================================
// MAIN WORKER SCHEDULER
// =====================================================================

type WorkerScheduler struct {
	// Core components
	db     *gorm.DB
	cron   *cron.Cron
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	config WorkerConfig

	// Job processing
	jobQueue   chan Job
	activeJobs int
	jobMutex   sync.RWMutex

	// Services
	services        *ServiceContainer
	jobProcessor    *JobProcessor
	resultProcessor *JobResultProcessor
	jobExecutor     JobExecutor
}

// =====================================================================
// SERVICE CONTAINER
// =====================================================================

type ServiceContainer struct {
	ReminderService       *ReminderService
	CleanupService        *CleanupService
	NotificationService   *NotificationService
	EnhancedNotifyService *EnhancedNotificationService
}

func NewServiceContainer(db *gorm.DB, emailService interfaces.EmailService, redisClient *redis.Client) *ServiceContainer {
	enhancedNotifyService := NewEnhancedNotificationService(db, redisClient, emailService)

	return &ServiceContainer{
		ReminderService:       NewReminderService(db, emailService, enhancedNotifyService),
		CleanupService:        NewCleanupService(db),
		NotificationService:   NewNotificationService(db),
		EnhancedNotifyService: enhancedNotifyService,
	}
}

// =====================================================================
// CONSTRUCTOR
// =====================================================================

func NewWorkerScheduler(db *gorm.DB, maxWorkers int, emailService interfaces.EmailService, redisClient *redis.Client) *WorkerScheduler {
	ctx, cancel := context.WithCancel(context.Background())

	config := WorkerConfig{
		MaxWorkers:      maxWorkers,
		QueueSize:       DefaultQueueSize,
		ShutdownTimeout: ShutdownTimeout,
	}

	// Configure cron with better options
	cronScheduler := cron.New(
		cron.WithParser(cron.NewParser(cron.Minute|cron.Hour|cron.Dom|cron.Month|cron.Dow|cron.Descriptor)),
		cron.WithChain(cron.Recover(cron.DefaultLogger)),
		cron.WithLogger(cron.DefaultLogger),
	)

	ws := &WorkerScheduler{
		db:       db,
		cron:     cronScheduler,
		ctx:      ctx,
		cancel:   cancel,
		config:   config,
		jobQueue: make(chan Job, config.QueueSize),
	}

	// Initialize services and processors
	ws.services = NewServiceContainer(db, emailService, redisClient)
	ws.jobProcessor = NewJobProcessor(db)
	ws.resultProcessor = NewJobResultProcessor(ws.jobProcessor, ws)
	ws.jobExecutor = NewJobExecutorImpl(ws.services)

	return ws
}

// =====================================================================
// LIFECYCLE MANAGEMENT
// =====================================================================

func (ws *WorkerScheduler) Start() error {
	log.Println("🚀 Starting Worker Scheduler...")

	if err := ws.startWorkers(); err != nil {
		return fmt.Errorf("failed to start workers: %w", err)
	}

	if err := ws.scheduleCronJobs(); err != nil {
		return fmt.Errorf("failed to schedule cron jobs: %w", err)
	}

	ws.cron.Start()
	ws.logStartupInfo()

	log.Printf("✅ Worker Scheduler started with %d workers", ws.config.MaxWorkers)
	return nil
}

func (ws *WorkerScheduler) Stop() {
	log.Println("🛑 Stopping Worker Scheduler...")

	// Create timeout context for graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), ws.config.ShutdownTimeout)
	defer shutdownCancel()

	// Stop cron scheduler
	cronCtx := ws.cron.Stop()

	// Wait for cron to stop or timeout
	select {
	case <-cronCtx.Done():
		log.Println("📅 Cron scheduler stopped")
	case <-shutdownCtx.Done():
		log.Println("⚠️ Cron scheduler shutdown timed out")
	}

	// Signal all goroutines to stop
	ws.cancel()

	// Close job queue
	close(ws.jobQueue)

	// Wait for workers to finish with timeout
	done := make(chan struct{})
	go func() {
		ws.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("✅ All workers stopped gracefully")
	case <-shutdownCtx.Done():
		log.Println("⚠️ Worker shutdown timed out")
	}

	log.Println("✅ Worker Scheduler stopped")
}

// =====================================================================
// WORKER MANAGEMENT
// =====================================================================

func (ws *WorkerScheduler) startWorkers() error {
	for i := 0; i < ws.config.MaxWorkers; i++ {
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
			ws.processJobWithTracking(job, id)

		case <-ws.ctx.Done():
			log.Printf("👷 Worker %d stopped by context", id)
			return
		}
	}
}

// =====================================================================
// JOB PROCESSING
// =====================================================================

func (ws *WorkerScheduler) processJobWithTracking(job Job, workerID int) {
	// Update active job count
	ws.updateActiveJobs(1)
	defer ws.updateActiveJobs(-1)

	// Create logger and record start
	logger := NewJobLogger(workerID)
	logger.LogStart(job)

	// Record job start in database
	if err := ws.jobProcessor.RecordJobStart(job); err != nil {
		log.Printf("⚠️ Failed to record job start: %v", err)
	}

	// Execute job with timeout
	err := ws.executeJobWithTimeout(job)

	// Process result
	ws.resultProcessor.ProcessResult(job, err, logger)
}

func (ws *WorkerScheduler) executeJobWithTimeout(job Job) error {
	// Create timeout context for job execution
	jobCtx, jobCancel := context.WithTimeout(ws.ctx, 5*time.Minute)
	defer jobCancel()

	// Execute job in a goroutine
	resultChan := make(chan error, 1)
	go func() {
		resultChan <- ws.jobExecutor.Execute(job)
	}()

	// Wait for job completion or timeout
	select {
	case err := <-resultChan:
		return err
	case <-jobCtx.Done():
		return fmt.Errorf("job execution timed out")
	}
}

func (ws *WorkerScheduler) updateActiveJobs(delta int) {
	ws.jobMutex.Lock()
	ws.activeJobs += delta
	ws.jobMutex.Unlock()
}

// =====================================================================
// JOB QUEUE MANAGEMENT
// =====================================================================

func (ws *WorkerScheduler) AddJob(job Job) {
	select {
	case <-ws.ctx.Done():
		log.Printf("⚠️ Cannot add job %s: scheduler is shutting down", job.ID)
		return
	default:
	}

	if ws.isQueueFull() {
		log.Printf("⚠️ Queue is full, dropping job: %s", job.ID)
		return
	}

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
	return ws.activeJobs >= ws.config.MaxWorkers*MaxQueueFullness
}

// =====================================================================
// CRON JOB SCHEDULING
// =====================================================================

func (ws *WorkerScheduler) getCronJobConfigs() []CronJobConfig {
	return []CronJobConfig{
		{Name: "process_notifications", Schedule: "* * * * *", JobType: "process_notifications", Priority: 1, Retries: 3},
		{Name: "booking_reminder", Schedule: "*/2 * * * *", JobType: "booking_reminder", Priority: 1, Retries: 3},
		{Name: "check_missed_bookings", Schedule: "*/15 * * * *", JobType: "check_missed_bookings", Priority: 2, Retries: 3},
		{Name: "handle_duplicate_bookings", Schedule: "*/30 * * * *", JobType: "handle_duplicate_bookings", Priority: 2, Retries: 3},
		{Name: "cleanup_old_data", Schedule: "0 2 * * *", JobType: "cleanup_old_data", Payload: map[string]interface{}{"days": 30}, Priority: 3, Retries: 2},
		{Name: "weekly_statistics", Schedule: "0 6 * * 0", JobType: "weekly_statistics", Priority: 2, Retries: 3},
	}
}

func (ws *WorkerScheduler) scheduleCronJobs() error {
	log.Println("📅 Scheduling cron jobs...")

	for _, config := range ws.getCronJobConfigs() {
		if err := ws.scheduleCronJob(config); err != nil {
			return fmt.Errorf("failed to schedule job %s: %w", config.Name, err)
		}
	}

	ws.scheduleTestJob()
	log.Println("📅 All cron jobs scheduled successfully")
	return nil
}

func (ws *WorkerScheduler) scheduleCronJob(config CronJobConfig) error {
	entryID, err := ws.cron.AddFunc(config.Schedule, func() {
		log.Printf("⏰ Cron triggered: %s", config.Name)
		ws.AddJob(Job{
			ID:         generateJobID(),
			Type:       config.JobType,
			Payload:    config.Payload,
			Priority:   config.Priority,
			MaxRetries: config.Retries,
			CreatedAt:  time.Now(),
		})
	})

	if err != nil {
		return err
	}

	log.Printf("✅ Scheduled %s (ID: %d) - %s", config.Name, entryID, config.Schedule)
	return nil
}

func (ws *WorkerScheduler) scheduleTestJob() {
	go func() {
		time.Sleep(TestStartupDelay)
		log.Println("⏰ Manual trigger: booking_reminder (test)")
		ws.AddJob(Job{
			ID:         generateJobID(),
			Type:       "booking_reminder",
			Priority:   1,
			MaxRetries: 3,
			CreatedAt:  time.Now(),
		})
	}()
}

// =====================================================================
// UTILITIES
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
		"max_workers":  ws.config.MaxWorkers,
		"queue_length": len(ws.jobQueue),
		"queue_cap":    cap(ws.jobQueue),
	}
}

// =====================================================================
// JOB EXECUTOR IMPLEMENTATION
// =====================================================================

type JobExecutorImpl struct {
	services *ServiceContainer
}

func NewJobExecutorImpl(services *ServiceContainer) *JobExecutorImpl {
	return &JobExecutorImpl{services: services}
}

func (je *JobExecutorImpl) Execute(job Job) error {
	switch job.Type {
	case "booking_reminder":
		return je.services.ReminderService.SendBookingReminders()
	case "check_missed_bookings":
		return je.services.ReminderService.CheckMissedBookings()
	case "handle_duplicate_bookings":
		return je.services.ReminderService.HandleDuplicateBookings()
	case "cleanup_old_data":
		days := je.extractCleanupDays(job.Payload)
		return je.services.CleanupService.CleanupOldData(days)
	case "weekly_statistics":
		return je.services.ReminderService.GenerateWeeklyStatistics()
	case "send_email_batch":
		return je.services.NotificationService.ProcessEmailBatch(job.Payload)
	case "send_email", "send_telegram", "send_sms":
		return je.services.EnhancedNotifyService.ProcessNotificationJob(job)
	default:
		return fmt.Errorf("unknown job type: %s", job.Type)
	}
}

func (je *JobExecutorImpl) extractCleanupDays(payload interface{}) int {
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

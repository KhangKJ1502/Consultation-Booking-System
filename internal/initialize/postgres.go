package initialize

import (
	"fmt"
	"log"

	"cbs_backend/global"
	entityLog "cbs_backend/internal/modules/activity_logs/entity"
	entityBooking "cbs_backend/internal/modules/bookings/entity"
	entityConsultation "cbs_backend/internal/modules/consultation_review/entity"
	entityExpert "cbs_backend/internal/modules/experts/entity"
	entityTemplate "cbs_backend/internal/modules/notification_template/entity"
	entityPayment "cbs_backend/internal/modules/payment_transactions/entity"
	entityPricing "cbs_backend/internal/modules/pricing_config/entity"
	entityNotification "cbs_backend/internal/modules/system_notification/entity"
	entitySystem "cbs_backend/internal/modules/system_setting/entity"
	entityUser "cbs_backend/internal/modules/users/entity"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitPostgres() *gorm.DB {
	cfg := global.ConfigConection

	if cfg == nil || cfg.PostgresCF == nil {
		log.Fatalf("❌ Config not loaded")
	}

	gormConfig := &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Info),
		DisableForeignKeyConstraintWhenMigrating: false,
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.PostgresCF.Host,
		cfg.PostgresCF.UserName,
		cfg.PostgresCF.Password,
		cfg.PostgresCF.DBName,
		cfg.PostgresCF.Port,
	)

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		log.Fatalf("failed to connect to PostgreSQL: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get underlying sql.DB: %v", err)
	}

	// Enable UUID extension
	if err := EnableUUIDExtension(db); err != nil {
		log.Fatalf("❌ Failed to enable UUID extension: %v", err)
	}

	// if err := MigrateDatabase(db); err != nil {
	// 	log.Fatalf("❌ Migration failed: %v", err)
	// }

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	global.DB = db
	log.Println("✅ PostgreSQL connected successfully")

	return db
}

func EnableUUIDExtension(db *gorm.DB) error {
	log.Println("🔧 Enabling UUID extension...")

	// Enable uuid-ossp extension for UUID generation
	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`).Error; err != nil {
		return fmt.Errorf("failed to enable uuid-ossp extension: %w", err)
	}

	log.Println("✅ UUID extension enabled successfully")
	return nil
}

func MigrateDatabase(db *gorm.DB) error {
	log.Println("🚀 Starting database migration...")

	// Migration order is critical due to foreign key dependencies
	// 1. Independent tables first (no foreign keys)
	independentTables := []interface{}{
		&entityUser.User{},
		&entitySystem.SystemSetting{},
		&entityTemplate.NotificationTemplate{},
	}

	userDependentTables := []interface{}{
		&entityExpert.ExpertProfile{},
		&entityUser.UserRefreshToken{},
		&entityUser.UserSession{},
		&entityNotification.SystemNotification{},
		&entityLog.ActivityLog{},
	}

	expertDependentTables := []interface{}{
		&entityPricing.PricingConfig{},
	}

	bookingRelatedTables := []interface{}{
		&entityBooking.ConsultationBooking{},
	}

	bookingDependentTables := []interface{}{
		&entityBooking.BookingStatusHistory{},
		&entityConsultation.ConsultationReview{},
		&entityPayment.PaymentTransaction{},
	}

	// Execute migrations in order
	migrationGroups := []struct {
		name   string
		tables []interface{}
	}{
		{"Independent tables", independentTables},
		{"User dependent tables", userDependentTables},
		{"Expert dependent tables", expertDependentTables},
		{"Booking related tables", bookingRelatedTables},
		{"Booking dependent tables", bookingDependentTables},
	}

	for _, group := range migrationGroups {
		log.Printf("📋 Migrating %s...", group.name)

		for _, table := range group.tables {
			if err := db.AutoMigrate(table); err != nil {
				return fmt.Errorf("failed to migrate %T: %w", table, err)
			}
			log.Printf("  ✅ Migrated %T", table)
		}
	}

	// Create indexes for better performance
	// if err := CreateIndexes(db); err != nil {
	// 	log.Printf("⚠️  Warning: Failed to create some indexes: %v", err)
	// }

	log.Println("✅ Database migrated successfully")
	return nil
}

// func CreateIndexes(db *gorm.DB) error {
// 	log.Println("📇 Creating database indexes...")

// 	indexes := []struct {
// 		table string
// 		index string
// 	}{
// 		// User table indexes
// 		{"tbl_users", "CREATE INDEX IF NOT EXISTS idx_users_email ON tbl_users(user_email);"},
// 		{"tbl_users", "CREATE INDEX IF NOT EXISTS idx_users_role ON tbl_users(user_role);"},
// 		{"tbl_users", "CREATE INDEX IF NOT EXISTS idx_users_active ON tbl_users(is_active);"},
// 		{"tbl_users", "CREATE INDEX IF NOT EXISTS idx_users_created_at ON tbl_users(user_created_at);"},

// 		// Expert profile indexes
// 		{"tbl_expert_profiles", "CREATE INDEX IF NOT EXISTS idx_expert_profiles_user_id ON tbl_expert_profiles(user_id);"},
// 		{"tbl_expert_profiles", "CREATE INDEX IF NOT EXISTS idx_expert_profiles_verified ON tbl_expert_profiles(is_verified);"},
// 		{"tbl_expert_profiles", "CREATE INDEX IF NOT EXISTS idx_expert_profiles_rating ON tbl_expert_profiles(average_rating);"},
// 		{"tbl_expert_profiles", "CREATE INDEX IF NOT EXISTS idx_expert_profiles_fee ON tbl_expert_profiles(consultation_fee);"},
// 		{"tbl_expert_profiles", "CREATE INDEX IF NOT EXISTS idx_expert_profiles_specialization ON tbl_expert_profiles USING GIN(specialization_list);"},

// 		// Booking indexes
// 		{"tbl_consultation_bookings", "CREATE INDEX IF NOT EXISTS idx_bookings_user_id ON tbl_consultation_bookings(user_id);"},
// 		{"tbl_consultation_bookings", "CREATE INDEX IF NOT EXISTS idx_bookings_expert_id ON tbl_consultation_bookings(expert_profile_id);"},
// 		{"tbl_consultation_bookings", "CREATE INDEX IF NOT EXISTS idx_bookings_status ON tbl_consultation_bookings(booking_status);"},
// 		{"tbl_consultation_bookings", "CREATE INDEX IF NOT EXISTS idx_bookings_datetime ON tbl_consultation_bookings(booking_datetime);"},
// 		{"tbl_consultation_bookings", "CREATE INDEX IF NOT EXISTS idx_bookings_type ON tbl_consultation_bookings(consultation_type);"},
// 		{"tbl_consultation_bookings", "CREATE INDEX IF NOT EXISTS idx_bookings_payment_status ON tbl_consultation_bookings(payment_status);"},

// 		// Working hours indexes
// 		{"tbl_expert_working_hours", "CREATE INDEX IF NOT EXISTS idx_working_hours_expert_id ON tbl_expert_working_hours(expert_profile_id);"},
// 		{"tbl_expert_working_hours", "CREATE INDEX IF NOT EXISTS idx_working_hours_day ON tbl_expert_working_hours(day_of_week);"},
// 		{"tbl_expert_working_hours", "CREATE INDEX IF NOT EXISTS idx_working_hours_active ON tbl_expert_working_hours(is_active);"},

// 		// Unavailable times indexes
// 		{"tbl_expert_unavailable_times", "CREATE INDEX IF NOT EXISTS idx_unavailable_times_expert_id ON tbl_expert_unavailable_times(expert_profile_id);"},
// 		{"tbl_expert_unavailable_times", "CREATE INDEX IF NOT EXISTS idx_unavailable_times_start ON tbl_expert_unavailable_times(unavailable_start_datetime);"},
// 		{"tbl_expert_unavailable_times", "CREATE INDEX IF NOT EXISTS idx_unavailable_times_end ON tbl_expert_unavailable_times(unavailable_end_datetime);"},

// 		// Review indexes
// 		{"tbl_consultation_reviews", "CREATE INDEX IF NOT EXISTS idx_reviews_expert_id ON tbl_consultation_reviews(expert_profile_id);"},
// 		{"tbl_consultation_reviews", "CREATE INDEX IF NOT EXISTS idx_reviews_reviewer_id ON tbl_consultation_reviews(reviewer_user_id);"},
// 		{"tbl_consultation_reviews", "CREATE INDEX IF NOT EXISTS idx_reviews_booking_id ON tbl_consultation_reviews(booking_id);"},
// 		{"tbl_consultation_reviews", "CREATE INDEX IF NOT EXISTS idx_reviews_rating ON tbl_consultation_reviews(rating_score);"},
// 		{"tbl_consultation_reviews", "CREATE INDEX IF NOT EXISTS idx_reviews_visible ON tbl_consultation_reviews(is_visible);"},

// 		// Payment transaction indexes
// 		{"tbl_payment_transactions", "CREATE INDEX IF NOT EXISTS idx_transactions_booking_id ON tbl_payment_transactions(booking_id);"},
// 		{"tbl_payment_transactions", "CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON tbl_payment_transactions(user_id);"},
// 		{"tbl_payment_transactions", "CREATE INDEX IF NOT EXISTS idx_transactions_expert_id ON tbl_payment_transactions(expert_profile_id);"},
// 		{"tbl_payment_transactions", "CREATE INDEX IF NOT EXISTS idx_transactions_status ON tbl_payment_transactions(transaction_status);"},
// 		{"tbl_payment_transactions", "CREATE INDEX IF NOT EXISTS idx_transactions_external_id ON tbl_payment_transactions(external_transaction_id);"},

// 		// Notification indexes
// 		{"tbl_system_notifications", "CREATE INDEX IF NOT EXISTS idx_notifications_recipient_id ON tbl_system_notifications(recipient_user_id);"},
// 		{"tbl_system_notifications", "CREATE INDEX IF NOT EXISTS idx_notifications_type ON tbl_system_notifications(notification_type);"},
// 		{"tbl_system_notifications", "CREATE INDEX IF NOT EXISTS idx_notifications_read ON tbl_system_notifications(is_read);"},
// 		{"tbl_system_notifications", "CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON tbl_system_notifications(notification_created_at);"},

// 		// Session indexes
// 		{"tbl_user_sessions", "CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON tbl_user_sessions(user_id);"},
// 		{"tbl_user_sessions", "CREATE INDEX IF NOT EXISTS idx_sessions_token ON tbl_user_sessions(session_token);"},
// 		{"tbl_user_sessions", "CREATE INDEX IF NOT EXISTS idx_sessions_active ON tbl_user_sessions(is_active);"},
// 		{"tbl_user_sessions", "CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON tbl_user_sessions(expires_at);"},

// 		// Refresh token indexes
// 		{"tbl_user_refresh_tokens", "CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON tbl_user_refresh_tokens(user_id);"},
// 		{"tbl_user_refresh_tokens", "CREATE INDEX IF NOT EXISTS idx_refresh_tokens_hash ON tbl_user_refresh_tokens(token_hash);"},
// 		{"tbl_user_refresh_tokens", "CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON tbl_user_refresh_tokens(expires_at);"},

// 		// Activity log indexes
// 		{"tbl_activity_logs", "CREATE INDEX IF NOT EXISTS idx_activity_logs_user_id ON tbl_activity_logs(user_id);"},
// 		{"tbl_activity_logs", "CREATE INDEX IF NOT EXISTS idx_activity_logs_action ON tbl_activity_logs(action_performed);"},
// 		{"tbl_activity_logs", "CREATE INDEX IF NOT EXISTS idx_activity_logs_table ON tbl_activity_logs(affected_table);"},
// 		{"tbl_activity_logs", "CREATE INDEX IF NOT EXISTS idx_activity_logs_created_at ON tbl_activity_logs(log_created_at);"},

// 		// System settings indexes
// 		{"tbl_system_settings", "CREATE INDEX IF NOT EXISTS idx_system_settings_key ON tbl_system_settings(setting_key);"},
// 		{"tbl_system_settings", "CREATE INDEX IF NOT EXISTS idx_system_settings_public ON tbl_system_settings(is_public);"},

// 		// Background jobs indexes
// 		{"tbl_background_jobs", "CREATE INDEX IF NOT EXISTS idx_background_jobs_type ON tbl_background_jobs(job_type);"},
// 		{"tbl_background_jobs", "CREATE INDEX IF NOT EXISTS idx_background_jobs_status ON tbl_background_jobs(job_status);"},
// 		{"tbl_background_jobs", "CREATE INDEX IF NOT EXISTS idx_background_jobs_scheduled_at ON tbl_background_jobs(scheduled_at);"},

// 		// Pricing config indexes
// 		{"tbl_pricing_configs", "CREATE INDEX IF NOT EXISTS idx_pricing_configs_expert_id ON tbl_pricing_configs(expert_profile_id);"},
// 		{"tbl_pricing_configs", "CREATE INDEX IF NOT EXISTS idx_pricing_configs_service_type ON tbl_pricing_configs(service_type);"},
// 		{"tbl_pricing_configs", "CREATE INDEX IF NOT EXISTS idx_pricing_configs_consultation_type ON tbl_pricing_configs(consultation_type);"},
// 		{"tbl_pricing_configs", "CREATE INDEX IF NOT EXISTS idx_pricing_configs_active ON tbl_pricing_configs(is_active);"},

// 		// Expert specialization indexes
// 		{"tbl_expert_specializations", "CREATE INDEX IF NOT EXISTS idx_expert_specializations_expert_id ON tbl_expert_specializations(expert_profile_id);"},
// 		{"tbl_expert_specializations", "CREATE INDEX IF NOT EXISTS idx_expert_specializations_name ON tbl_expert_specializations(specialization_name);"},
// 		{"tbl_expert_specializations", "CREATE INDEX IF NOT EXISTS idx_expert_specializations_primary ON tbl_expert_specializations(is_primary);"},

// 		// Booking status history indexes
// 		{"tbl_booking_status_history", "CREATE INDEX IF NOT EXISTS idx_booking_status_history_booking_id ON tbl_booking_status_history(booking_id);"},
// 		{"tbl_booking_status_history", "CREATE INDEX IF NOT EXISTS idx_booking_status_history_changed_by ON tbl_booking_status_history(changed_by_user_id);"},
// 		{"tbl_booking_status_history", "CREATE INDEX IF NOT EXISTS idx_booking_status_history_status ON tbl_booking_status_history(new_status);"},
// 		{"tbl_booking_status_history", "CREATE INDEX IF NOT EXISTS idx_booking_status_history_changed_at ON tbl_booking_status_history(status_changed_at);"},

// 		// Notification template indexes
// 		{"tbl_notification_templates", "CREATE INDEX IF NOT EXISTS idx_notification_templates_name ON tbl_notification_templates(template_name);"},
// 		{"tbl_notification_templates", "CREATE INDEX IF NOT EXISTS idx_notification_templates_type ON tbl_notification_templates(notification_type);"},
// 		{"tbl_notification_templates", "CREATE INDEX IF NOT EXISTS idx_notification_templates_active ON tbl_notification_templates(is_active);"},
// 	}

// 	for _, idx := range indexes {
// 		if err := db.Exec(idx.index).Error; err != nil {
// 			log.Printf("⚠️  Warning: Failed to create index on %s: %v", idx.table, err)
// 		}
// 	}

// 	log.Println("✅ Database indexes created successfully")
// 	return nil
// }

// // CreateDefaultData creates default system data
// func CreateDefaultData(db *gorm.DB) error {
// 	log.Println("📝 Creating default system data...")

// 	// Create default notification templates
// 	defaultTemplates := []models.NotificationTemplate{
// 		{
// 			TemplateName:     "booking_confirmed",
// 			NotificationType: "booking",
// 			TitleTemplate:    "Booking Confirmed",
// 			MessageTemplate:  "Your consultation booking has been confirmed for {{.datetime}}",
// 			TemplateVariables: models.JSONB{
// 				"datetime":    "string",
// 				"expert_name": "string",
// 			},
// 			IsActive: true,
// 		},
// 		{
// 			TemplateName:     "booking_reminder",
// 			NotificationType: "reminder",
// 			TitleTemplate:    "Upcoming Consultation",
// 			MessageTemplate:  "You have a consultation with {{.expert_name}} in 1 hour",
// 			TemplateVariables: models.JSONB{
// 				"expert_name": "string",
// 				"datetime":    "string",
// 			},
// 			IsActive: true,
// 		},
// 		{
// 			TemplateName:     "booking_cancelled",
// 			NotificationType: "booking",
// 			TitleTemplate:    "Booking Cancelled",
// 			MessageTemplate:  "Your consultation booking has been cancelled. Reason: {{.reason}}",
// 			TemplateVariables: models.JSONB{
// 				"reason":      "string",
// 				"expert_name": "string",
// 			},
// 			IsActive: true,
// 		},
// 	}

// 	for _, template := range defaultTemplates {
// 		var existing models.NotificationTemplate
// 		if err := db.Where("template_name = ?", template.TemplateName).First(&existing).Error; err != nil {
// 			if err == gorm.ErrRecordNotFound {
// 				if err := db.Create(&template).Error; err != nil {
// 					log.Printf("⚠️  Warning: Failed to create template %s: %v", template.TemplateName, err)
// 				}
// 			}
// 		}
// 	}

// 	// Create default system settings
// 	defaultSettings := []models.SystemSetting{
// 		{
// 			SettingKey:         "default_consultation_duration",
// 			SettingValue:       models.JSONB{"minutes": 60},
// 			SettingDescription: stringPtr("Default consultation duration in minutes"),
// 			IsPublic:           true,
// 		},
// 		{
// 			SettingKey:         "booking_reminder_hours",
// 			SettingValue:       models.JSONB{"hours": 1},
// 			SettingDescription: stringPtr("Hours before consultation to send reminder"),
// 			IsPublic:           false,
// 		},
// 		{
// 			SettingKey:         "max_booking_advance_days",
// 			SettingValue:       models.JSONB{"days": 30},
// 			SettingDescription: stringPtr("Maximum days in advance to allow booking"),
// 			IsPublic:           true,
// 		},
// 		{
// 			SettingKey:         "cancellation_policy_hours",
// 			SettingValue:       models.JSONB{"hours": 24},
// 			SettingDescription: stringPtr("Minimum hours before consultation to allow cancellation"),
// 			IsPublic:           true,
// 		},
// 	}

// 	for _, setting := range defaultSettings {
// 		var existing models.SystemSetting
// 		if err := db.Where("setting_key = ?", setting.SettingKey).First(&existing).Error; err != nil {
// 			if err == gorm.ErrRecordNotFound {
// 				if err := db.Create(&setting).Error; err != nil {
// 					log.Printf("⚠️  Warning: Failed to create setting %s: %v", setting.SettingKey, err)
// 				}
// 			}
// 		}
// 	}

// 	log.Println("✅ Default system data created successfully")
// 	return nil
// }

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}

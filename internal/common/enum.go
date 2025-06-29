package common

// Enums and Constants
const (
	// User roles
	UserRoleUser   = "user"
	UserRoleExpert = "expert"
	UserRoleAdmin  = "admin"

	// Gender options
	GenderMale   = "male"
	GenderFemale = "female"
	GenderOther  = "other"

	// Consultation types
	ConsultationTypeOnline  = "online"
	ConsultationTypeOffline = "offline"

	// Booking statuses
	BookingStatusPending   = "pending"
	BookingStatusConfirmed = "confirmed"
	BookingStatusRejected  = "rejected"
	BookingStatusCancelled = "cancelled"
	BookingStatusCompleted = "completed"
	BookingStatusMissed    = "missed"
	BookingStatusNoShow    = "no_show"

	// Payment statuses
	PaymentStatusPending  = "pending"
	PaymentStatusPaid     = "paid"
	PaymentStatusRefunded = "refunded"
	PaymentStatusFailed   = "failed"

	// Transaction statuses
	TransactionStatusPending    = "pending"
	TransactionStatusProcessing = "processing"
	TransactionStatusCompleted  = "completed"
	TransactionStatusFailed     = "failed"
	TransactionStatusRefunded   = "refunded"
	TransactionStatusCancelled  = "cancelled"

	// Job statuses
	JobStatusPending    = "pending"
	JobStatusProcessing = "processing"
	JobStatusCompleted  = "completed"
	JobStatusFailed     = "failed"
	JobStatusRetrying   = "retrying"

	// Days of week (0 = Sunday, 6 = Saturday)
	DaySunday    = 0
	DayMonday    = 1
	DayTuesday   = 2
	DayWednesday = 3
	DayThursday  = 4
	DayFriday    = 5
	DaySaturday  = 6

	// Default currency
	CurrencyVND = "VND"
	CurrencyUSD = "USD"
)

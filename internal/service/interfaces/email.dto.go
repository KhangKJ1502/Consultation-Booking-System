package interfaces

// Các data structures cho các loại email
type ConsultationBookingData struct {
	BookingID          string
	DoctorName         string
	DoctorSpecialty    string
	ConsultationDate   string
	ConsultationTime   string
	Duration           int    // minutes
	ConsultationType   string // "online" or "offline"
	Location           string // for offline consultations
	MeetingLink        string // for online consultations
	Amount             float64
	PaymentStatus      string
	BookingNotes       string
	CancellationPolicy string
}

type ConsultationReminderData struct {
	BookingID        string
	DoctorName       string
	ConsultationDate string
	ConsultationTime string
	MeetingLink      string
	Location         string
	ConsultationType string
	TimeUntil        string // "24 hours", "1 hour", etc.
}

type ConsultationCancellationData struct {
	BookingID         string
	DoctorName        string
	ConsultationDate  string
	ConsultationTime  string
	CancellationBy    string // "patient" or "doctor"
	CancellationNote  string
	RefundAmount      float64
	RefundProcessDays int
}

type ConsultationRescheduleData struct {
	BookingID           string
	DoctorName          string
	OldConsultationDate string
	OldConsultationTime string
	NewConsultationDate string
	NewConsultationTime string
	RescheduleBy        string
	RescheduleNote      string
}

type ConsultationNotificationData struct {
	BookingID        string
	PatientName      string
	PatientEmail     string
	PatientPhone     string
	ConsultationDate string
	ConsultationTime string
	PatientNotes     string
	ConsultationType string
}

type PaymentConfirmationData struct {
	BookingID        string
	TransactionID    string
	Amount           float64
	PaymentMethod    string
	PaymentDate      string
	DoctorName       string
	ConsultationDate string
	InvoiceURL       string
}

type PaymentFailedData struct {
	BookingID     string
	Amount        float64
	PaymentMethod string
	FailureReason string
	RetryURL      string
	BookingExpiry string
}

type RefundConfirmationData struct {
	BookingID      string
	TransactionID  string
	RefundAmount   float64
	RefundMethod   string
	ProcessingDays int
	RefundReason   string
}

type MaintenanceData struct {
	MaintenanceDate  string
	MaintenanceTime  string
	Duration         string
	AffectedServices []string
	ContactSupport   string
}

type NewsletterData struct {
	Subject        string
	ContentHTML    string
	UnsubscribeURL string
}

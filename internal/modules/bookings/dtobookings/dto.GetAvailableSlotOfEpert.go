package dtobookings

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// TimeOfDay represents time without date (for PostgreSQL TIME type)
type TimeOfDay struct {
	Hour   int
	Minute int
	Second int
}

// Scan implements the Scanner interface for database/sql
func (t *TimeOfDay) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case string:
		// Parse time string like "14:30:00" or "14:30"
		parsed, err := time.Parse("15:04:05", v)
		if err != nil {
			// Try parsing without seconds
			parsed, err = time.Parse("15:04", v)
			if err != nil {
				return fmt.Errorf("cannot parse time: %v", err)
			}
		}
		t.Hour = parsed.Hour()
		t.Minute = parsed.Minute()
		t.Second = parsed.Second()
		return nil
	case []byte:
		return t.Scan(string(v))
	default:
		return fmt.Errorf("cannot scan %T into TimeOfDay", value)
	}
}

// Value implements the driver Valuer interface
func (t TimeOfDay) Value() (driver.Value, error) {
	return fmt.Sprintf("%02d:%02d:%02d", t.Hour, t.Minute, t.Second), nil
}

// String returns string representation
func (t TimeOfDay) String() string {
	return fmt.Sprintf("%02d:%02d:%02d", t.Hour, t.Minute, t.Second)
}

// ToTime converts TimeOfDay to time.Time with given date
func (t TimeOfDay) ToTime(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(),
		t.Hour, t.Minute, t.Second, 0, date.Location())
}

// Request/Response DTOs
type GetAvailableSlotsRequest struct {
	ExpertProfileID     string    `form:"expert_profile_id" validate:"required,uuid"`
	FromDate            time.Time `form:"from_date" validate:"required" time_format:"2006-01-02T15:04:05Z07:00"`
	ToDate              time.Time `form:"to_date" validate:"required" time_format:"2006-01-02T15:04:05Z07:00"`
	SlotDurationMinutes int       `form:"slot_duration_minutes" validate:"min=15,max=240"`
}

type TimeSlot struct {
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
	DurationMinutes int       `json:"duration_minutes"`
	IsAvailable     bool      `json:"is_available"`
}

type GetAvailableSlotsResponse struct {
	ExpertProfileID string     `json:"expert_profile_id"`
	FromDate        time.Time  `json:"from_date"`
	ToDate          time.Time  `json:"to_date"`
	AvailableSlots  []TimeSlot `json:"available_slots"`
	TotalSlots      int        `json:"total_slots"`
	Message         string     `json:"message,omitempty"`
}

// Database row structs
type WorkingHourRow struct {
	DayOfWeek int       `json:"day_of_week"`
	StartTime TimeOfDay `json:"start_time"`
	EndTime   TimeOfDay `json:"end_time"`
}

type UnavailableTime struct {
	StartDatetime time.Time `json:"start_datetime"`
	EndDatetime   time.Time `json:"end_datetime"`
}

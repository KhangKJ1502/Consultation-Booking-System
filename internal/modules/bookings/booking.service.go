package bookings

import (
	"cbs_backend/internal/modules/bookings/dtobookings"
	"cbs_backend/utils/cache"
	"context"

	"github.com/bsm/redislock"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	iBookingService IBookings
)

func InitBookingService(db *gorm.DB, cache cache.BookingCache, logger *zap.Logger, redisLocker *redislock.Client) {
	iBookingService = NewBookingService(db, cache, logger, redisLocker)
}

func Booking() IBookings {
	if iBookingService == nil {
		panic("AuthService not initialized. Call InitBookingService(db,cache,logger) first.")
	}
	return iBookingService
}

type IBookings interface {
	CreateBooking(ctx context.Context, req dtobookings.CreateBookingRequest) (*dtobookings.CreateBookingResponse, error)
	GetUpcomingBookingsForExpert(ctx context.Context, req dtobookings.GetUpcomingBookingForExpertRequest) ([]*dtobookings.BookingResponse, error)
	CancelBooking(ctx context.Context, bookingID string, userID string) (*dtobookings.CancelResponse, error)
	ConfirmBooking(ctx context.Context, req dtobookings.ConfirmBooking) (*dtobookings.ConfirmBookingResponse, error)
	GetAvailableSlots(ctx context.Context, req dtobookings.GetAvailableSlotsRequest) (*dtobookings.GetAvailableSlotsResponse, error)
	UpdateBookingNotes(ctx context.Context, req dtobookings.UpdateBookingNotesRequest) (*dtobookings.UpdateBookingNotesResponse, error)
	GetBookingStatusHistory(ctx context.Context, req dtobookings.GetBookingStatusHistoryRequest) (*dtobookings.GetBookingStatusHistoryResponse, error)
}

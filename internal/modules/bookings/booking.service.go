package bookings

import (
	"cbs_backend/internal/modules/bookings/dtobookings"
	"cbs_backend/utils/cache"
	"context"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	iBookingService IBookings
)

func InitBookingService(db *gorm.DB, cache cache.BookingCache, logger *zap.Logger) {
	iBookingService = NewBookingService(db, cache, logger)
}

func Booking() IBookings {
	if iBookingService == nil {
		panic("AuthService not initialized. Call InitBookingService(db,cache,logger) first.")
	}
	return iBookingService
}

type IBookings interface {
	CreateBooking(ctx context.Context, req dtobookings.CreateBookingRequest) (*dtobookings.CreateBookingResponse, error)
}

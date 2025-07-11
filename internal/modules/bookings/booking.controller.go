package bookings

import (
	"cbs_backend/internal/modules/bookings/dtobookings"
	"cbs_backend/pkg/response"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type BookingController struct {
	Logger *zap.Logger // Thiếu field này
}

func NewBookingController(logger *zap.Logger) *BookingController {
	return &BookingController{Logger: logger}
}

func (bc *BookingController) CreateBooking(c *gin.Context) (res interface{}, err error) {
	var req dtobookings.CreateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bc.Logger.Error("Invalid create booking request", zap.Error(err))
		return nil, response.NewAPIError(http.StatusInternalServerError, "Invalid create booking request", err)
	}

	resp, err := Booking().CreateBooking(context.Background(), req)
	if err != nil {
		bc.Logger.Error("Create booking failed", zap.Error(err))
		return nil, response.NewAPIError(http.StatusInternalServerError, "Create booking failed", err)
	}

	return resp, nil
}
func (bc *BookingController) GetUpcomingBookingsForExpert(c *gin.Context) (res interface{}, err error) {
	var req dtobookings.GetUpcomingBookingForExpertRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		bc.Logger.Error("Invalid get upcoming bookings request", zap.Error(err))
		return nil, response.NewAPIError(http.StatusInternalServerError, "Invalid get upcoming bookings request", err)
	}

	bookings, err := Booking().GetUpcomingBookingsForExpert(context.Background(), req)
	if err != nil {
		bc.Logger.Error("Get upcoming bookings failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return nil, response.NewAPIError(http.StatusInternalServerError, "Get upcoming bookings failed", err)
	}

	return bookings, nil
}
func (bc *BookingController) CancelBooking(c *gin.Context) (res interface{}, err error) {
	bookingID := c.Param("bookingID")
	// userID := c.GetHeader("userID")
	userID := "e2a5ee96-4ffb-40e9-96be-12fff3576ff7"
	if bookingID == "" || userID == "" {
		bc.Logger.Error("Missing bookingID or userID", zap.Error(err))
		return nil, response.NewAPIError(http.StatusBadRequest, "Missing bookingID or userID", err)
	}

	resp, err := Booking().CancelBooking(c, bookingID, userID)
	if err != nil {
		bc.Logger.Error("Cancel booking failed", zap.Error(err))
		return nil, response.NewAPIError(http.StatusBadRequest, "Cancel booking failed", err)
	}

	return resp, nil
}
func (bc *BookingController) ConfirmBooking(c *gin.Context) (res interface{}, err error) {
	var req dtobookings.ConfirmBooking
	if err := c.ShouldBindJSON(&req); err != nil {
		bc.Logger.Error("Invalid confirm booking request", zap.Error(err))
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid confirm booking request", err)
	}

	resp, err := Booking().ConfirmBooking(context.Background(), req)
	if err != nil {
		bc.Logger.Error("Confirm booking failed", zap.Error(err))
		return nil, response.NewAPIError(http.StatusBadRequest, "Confirm booking failed", err)
	}

	return resp, nil
}
func (bc *BookingController) GetAvailableSlots(c *gin.Context) (res interface{}, err error) {
	var req dtobookings.GetAvailableSlotsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		bc.Logger.Error("Invalid get available slots request", zap.Error(err))
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid get available slots request", err)
	}

	resp, err := Booking().GetAvailableSlots(context.Background(), req)
	if err != nil {
		bc.Logger.Error("Get available slots failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return nil, response.NewAPIError(http.StatusBadRequest, "Get available slots failed", err)
	}
	return resp, nil
}
func (bc *BookingController) UpdateBookingNotes(c *gin.Context) (res interface{}, err error) {
	var req dtobookings.UpdateBookingNotesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bc.Logger.Error("Invalid update booking notes request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid update booking notes request", err)
	}

	resp, err := Booking().UpdateBookingNotes(context.Background(), req)
	if err != nil {
		bc.Logger.Error("Update booking notes failed", zap.Error(err))
		return nil, response.NewAPIError(http.StatusBadRequest, "Update booking notes failed", err)
	}

	return resp, nil
}
func (bc *BookingController) GetBookingStatusHistory(c *gin.Context) (res interface{}, err error) {
	var req dtobookings.GetBookingStatusHistoryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		bc.Logger.Error("Invalid get booking status history request", zap.Error(err))
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid get booking status history request", err)
	}

	resp, err := Booking().GetBookingStatusHistory(context.Background(), req)
	if err != nil {
		bc.Logger.Error("Get booking status history failed", zap.Error(err))
		return nil, response.NewAPIError(http.StatusBadRequest, "Get booking status history failed", err)
	}
	return resp, nil
}

func (bc *BookingController) GetBookingByID(c *gin.Context) (res interface{}, err error) {
	var req dtobookings.GetBookingByIDRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		bc.Logger.Error("Invalid get booking by ID request", zap.Error(err))
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid get booking by ID request", err)
	}

	resp, err := Booking().GetBookingByID(context.Background(), req)
	if err != nil {
		bc.Logger.Error("Get booking by ID failed", zap.Error(err))
		return nil, response.NewAPIError(http.StatusInternalServerError, "Get booking by ID failed", err)
	}

	return resp, nil
}

func (bc *BookingController) GetUserBookingHistory(c *gin.Context) (res interface{}, err error) {
	var req dtobookings.GetUserBookingHistoryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		bc.Logger.Error("Invalid get user booking history request", zap.Error(err))
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid get user booking history request", err)
	}

	resp, err := Booking().GetUserBookingHistory(context.Background(), req)
	if err != nil {
		bc.Logger.Error("Get user booking history failed", zap.Error(err))
		return nil, response.NewAPIError(http.StatusInternalServerError, "Get user booking history failed", err)
	}

	return resp, nil
}

func (bc *BookingController) RescheduleBooking(c *gin.Context) (res interface{}, err error) {
	var req dtobookings.RescheduleBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bc.Logger.Error("Invalid reschedule booking request", zap.Error(err))
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid reschedule booking request", err)
	}

	resp, err := Booking().RescheduleBooking(context.Background(), req)
	if err != nil {
		bc.Logger.Error("Reschedule booking failed", zap.Error(err))
		return nil, response.NewAPIError(http.StatusBadRequest, "Reschedule booking failed", err)
	}

	return resp, nil
}

func (bc *BookingController) CompleteBooking(c *gin.Context) (res interface{}, err error) {
	var req dtobookings.CompleteBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bc.Logger.Error("Invalid complete booking request", zap.Error(err))
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid complete booking request", err)
	}

	resp, err := Booking().CompleteBooking(context.Background(), req)
	if err != nil {
		bc.Logger.Error("Complete booking failed", zap.Error(err))
		return nil, response.NewAPIError(http.StatusBadRequest, "Complete booking failed", err)
	}

	return resp, nil
}

func (bc *BookingController) GetBookingStats(c *gin.Context) (res interface{}, err error) {
	var req dtobookings.GetBookingStatsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		bc.Logger.Error("Invalid get booking stats request", zap.Error(err))
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid get booking stats request", err)
	}

	resp, err := Booking().GetBookingStats(context.Background(), req)
	if err != nil {
		bc.Logger.Error("Get booking stats failed", zap.Error(err))
		return nil, response.NewAPIError(http.StatusInternalServerError, "Get booking stats failed", err)
	}

	return resp, nil
}

func (bc *BookingController) SearchBookings(c *gin.Context) (res interface{}, err error) {
	var req dtobookings.SearchBookingsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		bc.Logger.Error("Invalid search bookings request", zap.Error(err))
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid search bookings request", err)
	}

	resp, err := Booking().SearchBookings(context.Background(), req)
	if err != nil {
		bc.Logger.Error("Search bookings failed", zap.Error(err))
		return nil, response.NewAPIError(http.StatusInternalServerError, "Search bookings failed", err)
	}

	return resp, nil
}

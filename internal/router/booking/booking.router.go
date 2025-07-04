package booking

import (
	"cbs_backend/global"
	"cbs_backend/pkg/response"

	"cbs_backend/internal/middleware"
	PkgBooking "cbs_backend/internal/modules/bookings"
	"cbs_backend/internal/modules/users"

	"github.com/gin-gonic/gin"
)

type BookingRouter struct{}

func (br *BookingRouter) InitBookingRouter(router *gin.RouterGroup) {
	bookingCtr := PkgBooking.NewBookingController(global.Log)

	// Public group: không cần đăng nhập
	bookingPublic := router.Group("/booking/v1")
	{
		bookingPublic.GET("/available-slots", response.Wrap(bookingCtr.GetAvailableSlots))
		// Nếu /upcoming chỉ trả thông tin public, có thể để ở đây.
		// bookingPublic.GET("/upcoming", response.Wrap(bookingCtr.GetUpcomingBookingsForExpert))
	}

	// Private group: cần đăng nhập
	bookingPrivate := router.Group("/booking/v2")
	bookingPrivate.Use(middleware.AuthMiddleware(users.User()))
	{
		bookingPrivate.POST("/", response.Wrap(bookingCtr.CreateBooking))
		bookingPrivate.GET("/upcoming", response.Wrap(bookingCtr.GetUpcomingBookingsForExpert))
		bookingPrivate.POST("/cancel/:bookingID", response.Wrap(bookingCtr.CancelBooking))
		bookingPrivate.POST("/confirm", response.Wrap(bookingCtr.ConfirmBooking))
		bookingPrivate.POST("/update-notes", response.Wrap(bookingCtr.UpdateBookingNotes))
		bookingPrivate.GET("/status-history", response.Wrap(bookingCtr.GetBookingStatusHistory))
	}
}

package router

import (
	"cbs_backend/internal/router/booking"
	"cbs_backend/internal/router/dashboard"
	"cbs_backend/internal/router/expert"
	"cbs_backend/internal/router/user"
)

type RouterGroup struct {
	User      user.RouterUserGroup
	Expert    expert.RouterExpertGroup
	Booking   booking.RouterBookingGroup
	Dashboard dashboard.RouterDashBoardGroup
}

var RouterGroupApp = new(RouterGroup)

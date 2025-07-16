package dashboard

import (
	"cbs_backend/global"
	"cbs_backend/internal/modules/dashboard/dtodashboard"
	"cbs_backend/pkg/response"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type DashboardController struct {
	Logger *zap.Logger
}

func NewDashboardController() *DashboardController {
	logger := global.Log
	return &DashboardController{Logger: logger}
}

func (dc *DashboardController) GetBookingStats(c *gin.Context) (res interface{}, err error) {
	var req dtodashboard.BookingStatsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dc.Logger.Error("Invalid get booking stats request", zap.Error(err))
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid get booking stats request", err)
	}

	resp, err := Dashboard().GetBookingStats(context.Background(), req)
	if err != nil {
		dc.Logger.Error("Get booking stats failed", zap.Error(err))
		return nil, response.NewAPIError(http.StatusInternalServerError, "Get booking stats failed", err)
	}

	return resp, nil
}

func (dc *DashboardController) GetSystemOverview(c *gin.Context) (res interface{}, err error) {
	resp, err := Dashboard().GetSystemOverview(context.Background())
	if err != nil {
		dc.Logger.Error("Get system overview failed", zap.Error(err))
		return nil, response.NewAPIError(http.StatusInternalServerError, "Get system overview failed", err)
	}

	return resp, nil
}

func (dc *DashboardController) GetRevenueReport(c *gin.Context) (res interface{}, err error) {
	var req dtodashboard.RevenueReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dc.Logger.Error("Invalid get revenue report request", zap.Error(err))
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid get revenue report request", err)
	}

	resp, err := Dashboard().GetRevenueReport(context.Background(), req)
	if err != nil {
		dc.Logger.Error("Get revenue report failed", zap.Error(err))
		return nil, response.NewAPIError(http.StatusInternalServerError, "Get revenue report failed", err)
	}

	return resp, nil
}

func (dc *DashboardController) GetExpertPerformance(c *gin.Context) (res interface{}, err error) {
	expertId := c.Param("expertId")
	if expertId == "" {
		dc.Logger.Error("Missing expertId parameter")
		return nil, response.NewAPIError(http.StatusBadRequest, "Missing expertId parameter", nil)
	}

	resp, err := Dashboard().GetExpertPerformance(context.Background(), expertId)
	if err != nil {
		dc.Logger.Error("Get expert performance failed", zap.Error(err))
		return nil, response.NewAPIError(http.StatusInternalServerError, "Get expert performance failed", err)
	}

	return resp, nil
}

package dashboard

import (
	"cbs_backend/internal/modules/dashboard/dtodashboard"
	"context"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	iDashboardService IDashboard
)

func InitDashboardService(db *gorm.DB, logger *zap.Logger) {
	iDashboardService = NewDashBoardService(db, logger)
}

func Dashboard() IDashboard {
	if iDashboardService == nil {
		panic("DashboardService not initialized. Call InitDashboardService(db, logger) first.")
	}
	return iDashboardService
}

type IDashboard interface {
	GetBookingStats(ctx context.Context, req dtodashboard.BookingStatsRequest) (res dtodashboard.BookingStatsResponse, err error)
	GetSystemOverview(ctx context.Context) (res dtodashboard.SystemOverviewResponse, err error)
	GetRevenueReport(ctx context.Context, req dtodashboard.RevenueReportRequest) (res dtodashboard.RevenueReportResponse, err error)
	GetExpertPerformance(ctx context.Context, expertId string) (res dtodashboard.ExpertPerformanceResponse, err error)
}

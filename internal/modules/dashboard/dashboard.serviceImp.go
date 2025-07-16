package dashboard

import (
	"cbs_backend/internal/modules/dashboard/dtodashboard"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type DashboardService struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewDashBoardService(db *gorm.DB, logger *zap.Logger) *DashboardService {
	return &DashboardService{
		db:     db,
		logger: logger,
	}
}

func (dbs *DashboardService) GetBookingStats(ctx context.Context, req dtodashboard.BookingStatsRequest) (res dtodashboard.BookingStatsResponse, err error) {
	dbs.logger.Info("📊 Getting booking stats", zap.Any("request", req))

	var (
		count   int64
		revenue float64
	)

	// Bắt đầu query cơ bản
	query := dbs.db.WithContext(ctx).Table("tbl_consultation_bookings AS cb").
		Joins("LEFT JOIN tbl_payment_transactions AS pt ON cb.booking_id = pt.booking_id")

	// Lọc theo ngày nếu có
	if req.DateFrom != nil {
		query = query.Where("cb.booking_created_at >= ?", req.DateFrom)
	}
	if req.DateTo != nil {
		query = query.Where("cb.booking_created_at <= ?", req.DateTo)
	}

	// Lọc theo expert_id nếu hợp lệ
	if req.ExpertID != nil && *req.ExpertID != "" {
		// Kiểm tra chuỗi expertID có phải là UUID hợp lệ không
		if _, parseErr := uuid.Parse(*req.ExpertID); parseErr == nil {
			query = query.Where("cb.expert_profile_id = ?", *req.ExpertID)
		} else {
			dbs.logger.Warn("❌ Invalid UUID for ExpertID", zap.String("expert_id", *req.ExpertID))
			return res, fmt.Errorf("invalid expert_id UUID format")
		}
	}

	// Lọc theo status nếu có
	if req.Status != nil && *req.Status != "" {
		query = query.Where("cb.booking_status = ?", *req.Status)
	}

	// Đếm số lượng booking
	if err = query.Count(&count).Error; err != nil {
		dbs.logger.Error("❌ Failed to count bookings", zap.Error(err))
		return res, fmt.Errorf("get booking count failed: %w", err)
	}

	// Tính tổng doanh thu từ các giao dịch thành công
	revenueQuery := query.Session(&gorm.Session{}) // tránh ảnh hưởng bởi Count()
	revenueQuery = revenueQuery.Where("pt.transaction_status = ?", "completed")
	if err = revenueQuery.Select("COALESCE(SUM(pt.amount), 0)").Scan(&revenue).Error; err != nil {
		dbs.logger.Error("❌ Failed to calculate revenue", zap.Error(err))
		return res, fmt.Errorf("get booking revenue failed: %w", err)
	}

	// Định dạng thời gian cho phản hồi
	period := "All time"
	if req.DateFrom != nil && req.DateTo != nil {
		period = fmt.Sprintf("%s to %s", req.DateFrom.Format("2006-01-02"), req.DateTo.Format("2006-01-02"))
	}

	// Lấy trạng thái hiển thị
	status := "all"
	if req.Status != nil && *req.Status != "" {
		status = *req.Status
	}

	// Gán kết quả vào response
	res = dtodashboard.BookingStatsResponse{
		Period:  period,
		Count:   count,
		Revenue: revenue,
		Status:  status,
	}

	return res, nil
}

func (dbs *DashboardService) GetSystemOverview(ctx context.Context) (res dtodashboard.SystemOverviewResponse, err error) {
	dbs.logger.Info("Getting system overview")

	var totalBookings int64
	var pendingBookings int64
	var confirmedBookings int64
	var completedBookings int64
	var cancelledBookings int64
	var activeExperts int64
	var activeUsers int64
	var totalRevenue float64

	// Get total bookings
	if err = dbs.db.WithContext(ctx).Table("tbl_consultation_bookings").Count(&totalBookings).Error; err != nil {
		dbs.logger.Error("Failed to get total bookings", zap.Error(err))
		return res, err
	}

	// Get pending bookings
	if err = dbs.db.WithContext(ctx).Table("tbl_consultation_bookings").Where("booking_status = ?", "pending").Count(&pendingBookings).Error; err != nil {
		dbs.logger.Error("Failed to get pending bookings", zap.Error(err))
		return res, err
	}

	// Get confirmed bookings
	if err = dbs.db.WithContext(ctx).Table("tbl_consultation_bookings").Where("booking_status = ?", "confirmed").Count(&confirmedBookings).Error; err != nil {
		dbs.logger.Error("Failed to get confirmed bookings", zap.Error(err))
		return res, err
	}

	// Get completed bookings
	if err = dbs.db.WithContext(ctx).Table("tbl_consultation_bookings").Where("booking_status = ?", "completed").Count(&completedBookings).Error; err != nil {
		dbs.logger.Error("Failed to get completed bookings", zap.Error(err))
		return res, err
	}

	// Get cancelled bookings
	if err = dbs.db.WithContext(ctx).Table("tbl_consultation_bookings").Where("booking_status = ?", "cancelled").Count(&cancelledBookings).Error; err != nil {
		dbs.logger.Error("Failed to get cancelled bookings", zap.Error(err))
		return res, err
	}

	// Get active experts
	if err = dbs.db.WithContext(ctx).Table("tbl_expert_profiles").
		Joins("JOIN tbl_users ON tbl_expert_profiles.user_id = tbl_users.user_id").
		Where("tbl_users.is_active = ? AND tbl_expert_profiles.is_verified = ?", true, true).
		Count(&activeExperts).Error; err != nil {
		dbs.logger.Error("Failed to get active experts", zap.Error(err))
		return res, err
	}

	// Get active users
	if err = dbs.db.WithContext(ctx).Table("tbl_users").Where("is_active = ?", true).Count(&activeUsers).Error; err != nil {
		dbs.logger.Error("Failed to get active users", zap.Error(err))
		return res, err
	}

	// Get total revenue from completed transactions
	if err = dbs.db.WithContext(ctx).Table("tbl_payment_transactions").
		Select("COALESCE(SUM(amount), 0)").
		Where("transaction_status = ?", "completed").
		Scan(&totalRevenue).Error; err != nil {
		dbs.logger.Error("Failed to get total revenue", zap.Error(err))
		return res, err
	}

	// Calculate success rate
	var successRate float64
	if totalBookings > 0 {
		successRate = (float64(completedBookings) / float64(totalBookings)) * 100
	}

	res = dtodashboard.SystemOverviewResponse{
		TotalBookings:     totalBookings,
		PendingBookings:   pendingBookings,
		ConfirmedBookings: confirmedBookings,
		CompletedBookings: completedBookings,
		CancelledBookings: cancelledBookings,
		ActiveExperts:     activeExperts,
		ActiveUsers:       activeUsers,
		TotalRevenue:      totalRevenue,
		SuccessRate:       successRate,
	}

	return res, nil
}

func (dbs *DashboardService) GetRevenueReport(ctx context.Context, req dtodashboard.RevenueReportRequest) (res dtodashboard.RevenueReportResponse, err error) {
	dbs.logger.Info("Getting revenue report", zap.Any("request", req))

	var revenue float64
	var bookingCount int64

	// Get revenue for the period
	if err = dbs.db.WithContext(ctx).Table("tbl_payment_transactions").
		Select("COALESCE(SUM(amount), 0)").
		Where("transaction_status = ? AND transaction_created_at >= ? AND transaction_created_at <= ?", "completed", req.DateFrom, req.DateTo).
		Scan(&revenue).Error; err != nil {
		dbs.logger.Error("Failed to get revenue", zap.Error(err))
		return res, err
	}

	// Get booking count for the period
	if err = dbs.db.WithContext(ctx).Table("tbl_consultation_bookings").
		Where("booking_created_at >= ? AND booking_created_at <= ?", req.DateFrom, req.DateTo).
		Count(&bookingCount).Error; err != nil {
		dbs.logger.Error("Failed to get booking count", zap.Error(err))
		return res, err
	}

	// Calculate growth percentage (compared to previous period)
	var previousRevenue float64
	periodDuration := req.DateTo.Sub(req.DateFrom)
	previousStartDate := req.DateFrom.Add(-periodDuration)
	previousEndDate := req.DateFrom

	if err = dbs.db.WithContext(ctx).Table("tbl_payment_transactions").
		Select("COALESCE(SUM(amount), 0)").
		Where("transaction_status = ? AND transaction_created_at >= ? AND transaction_created_at < ?", "completed", previousStartDate, previousEndDate).
		Scan(&previousRevenue).Error; err != nil {
		dbs.logger.Error("Failed to get previous revenue", zap.Error(err))
		return res, err
	}

	var growth float64
	if previousRevenue > 0 {
		growth = ((revenue - previousRevenue) / previousRevenue) * 100
	}

	// Format period
	period := req.DateFrom.Format("2006-01-02") + " to " + req.DateTo.Format("2006-01-02")

	res = dtodashboard.RevenueReportResponse{
		Period:       period,
		Revenue:      revenue,
		BookingCount: bookingCount,
		Growth:       growth,
	}

	return res, nil
}

func (dbs *DashboardService) GetExpertPerformance(ctx context.Context, expertId string) (res dtodashboard.ExpertPerformanceResponse, err error) {
	dbs.logger.Info("Getting expert performance", zap.String("expertId", expertId))

	// Parse UUID
	expertUUID, err := uuid.Parse(expertId)
	if err != nil {
		dbs.logger.Error("Invalid expert ID format", zap.String("expertId", expertId), zap.Error(err))
		return res, fmt.Errorf("invalid expert ID format: %w", err)
	}

	// Struct tạm để map kết quả
	var expertProfile struct {
		ExpertProfileID string
		ExpertName      string
	}

	// Lấy expert profile ID và tên
	if err = dbs.db.WithContext(ctx).Table("tbl_expert_profiles AS ep").
		Select("ep.expert_profile_id, u.full_name AS expert_name").
		Joins("JOIN tbl_users u ON ep.user_id = u.user_id").
		Where("ep.expert_profile_id = ?", expertUUID).
		First(&expertProfile).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			dbs.logger.Warn("Expert not found", zap.String("expertId", expertId))
			return res, fmt.Errorf("expert not found")
		}
		dbs.logger.Error("Failed to get expert profile", zap.Error(err))
		return res, err
	}

	expertProfileId := expertProfile.ExpertProfileID
	expertName := expertProfile.ExpertName

	var (
		totalBookings     int64
		completedBookings int64
		cancelledBookings int64
		averageRating     float64
		revenue           float64
	)

	// Total bookings
	if err = dbs.db.WithContext(ctx).Table("tbl_consultation_bookings").
		Where("expert_profile_id = ?", expertProfileId).
		Count(&totalBookings).Error; err != nil {
		dbs.logger.Error("Failed to get total bookings", zap.Error(err))
		return res, err
	}

	// Completed bookings
	if err = dbs.db.WithContext(ctx).Table("tbl_consultation_bookings").
		Where("expert_profile_id = ? AND booking_status = ?", expertProfileId, "completed").
		Count(&completedBookings).Error; err != nil {
		dbs.logger.Error("Failed to get completed bookings", zap.Error(err))
		return res, err
	}

	// Cancelled bookings
	if err = dbs.db.WithContext(ctx).Table("tbl_consultation_bookings").
		Where("expert_profile_id = ? AND booking_status = ?", expertProfileId, "cancelled").
		Count(&cancelledBookings).Error; err != nil {
		dbs.logger.Error("Failed to get cancelled bookings", zap.Error(err))
		return res, err
	}

	// Average rating
	if err = dbs.db.WithContext(ctx).Table("tbl_consultation_reviews").
		Where("expert_profile_id = ?", expertProfileId).
		Select("COALESCE(AVG(rating_score), 0)").Scan(&averageRating).Error; err != nil {
		dbs.logger.Error("Failed to get average rating", zap.Error(err))
		return res, err
	}

	// Revenue
	if err = dbs.db.WithContext(ctx).Table("tbl_payment_transactions").
		Where("expert_profile_id = ? AND transaction_status = ?", expertProfileId, "completed").
		Select("COALESCE(SUM(amount), 0)").Scan(&revenue).Error; err != nil {
		dbs.logger.Error("Failed to get revenue", zap.Error(err))
		return res, err
	}

	// Success rate
	var successRate float64
	if totalBookings > 0 {
		successRate = (float64(completedBookings) / float64(totalBookings)) * 100
	}

	// Trả kết quả
	res = dtodashboard.ExpertPerformanceResponse{
		ExpertID:          expertUUID.String(),
		ExpertName:        expertName,
		TotalBookings:     totalBookings,
		CompletedBookings: completedBookings,
		CancelledBookings: cancelledBookings,
		Revenue:           revenue,
		AverageRating:     averageRating,
		SuccessRate:       successRate,
	}

	return res, nil
}

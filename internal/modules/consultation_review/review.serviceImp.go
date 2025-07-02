package consultationreview

import (
	entityBooking "cbs_backend/internal/modules/bookings/entity"
	"cbs_backend/internal/modules/consultation_review/dtoreviews"
	"cbs_backend/internal/modules/consultation_review/entity"
	entityReview "cbs_backend/internal/modules/consultation_review/entity"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type reviewService struct {
	db *gorm.DB
	// cache  utils.BookingCache
	logger *zap.Logger
	// helper *utilshelper.HelperBooking
}

func NewReviewService(db *gorm.DB, logger *zap.Logger) *reviewService {
	return &reviewService{
		db:     db,
		logger: logger,
		// helper: helper.NewHelperBooking(db),
	}
}
func (rs *reviewService) CreateReview(ctx context.Context, req dtoreviews.CreateReviewRequest) (*dtoreviews.CreateReviewResponse, error) {
	// Validate IDs
	bookingID, err := uuid.Parse(req.BookingID)
	if err != nil {
		return nil, fmt.Errorf("invalid booking ID format: %w", err)
	}

	reviewerID, err := uuid.Parse(req.ReviewerUserID)
	if err != nil {
		return nil, fmt.Errorf("invalid reviewer user ID format: %w", err)
	}

	// Validate rating score
	if req.RatingScore < 1 || req.RatingScore > 5 {
		return nil, fmt.Errorf("rating score must be between 1 and 5")
	}

	// Get booking details
	var booking entityBooking.ConsultationBooking
	if err := rs.db.WithContext(ctx).First(&booking, "booking_id = ?", bookingID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("booking not found")
		}
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}

	// Authorization - chỉ user đặt lịch mới được review
	if booking.UserID != reviewerID {
		return nil, fmt.Errorf("unauthorized: only the booking owner can create a review")
	}

	// Chỉ cho phép review khi booking đã completed
	if booking.BookingStatus != "completed" {
		return nil, fmt.Errorf("can only review completed bookings")
	}

	// Check if review already exists
	var existingReview entity.ConsultationReview
	err = rs.db.WithContext(ctx).
		Where("booking_id = ? AND reviewer_user_id = ?", bookingID, reviewerID).
		First(&existingReview).Error

	if err == nil {
		return nil, fmt.Errorf("review already exists for this booking")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check existing review: %w", err)
	}

	// Create review
	newReview := &entityReview.ConsultationReview{
		BookingID:       bookingID,
		ReviewerUserID:  reviewerID,
		ExpertProfileID: booking.ExpertProfileID,
		RatingScore:     req.RatingScore,
		ReviewComment:   &req.ReviewComment,
		IsAnonymous:     req.IsAnonymous,
		IsVisible:       true, // Default to visible
		ReviewCreatedAt: time.Now(),
		ReviewUpdatedAt: time.Now(),
	}

	// Start transaction
	tx := rs.db.WithContext(ctx).Begin()

	if err := tx.Create(newReview).Error; err != nil {
		tx.Rollback()
		rs.logger.Error("Failed to create review", zap.Error(err))
		return nil, fmt.Errorf("failed to create review: %w", err)
	}

	// Update expert's average rating
	if err := rs.updateExpertRating(tx, booking.ExpertProfileID); err != nil {
		tx.Rollback()
		rs.logger.Error("Failed to update expert rating", zap.Error(err))
		return nil, fmt.Errorf("failed to update expert rating: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		rs.logger.Error("Failed to commit review transaction", zap.Error(err))
		return nil, fmt.Errorf("failed to commit review creation: %w", err)
	}

	response := &dtoreviews.CreateReviewResponse{
		ReviewID:        newReview.ReviewID.String(),
		BookingID:       newReview.BookingID.String(),
		ReviewerUserID:  newReview.ReviewerUserID.String(),
		ExpertProfileID: newReview.ExpertProfileID.String(),
		RatingScore:     newReview.RatingScore,
		ReviewComment:   *newReview.ReviewComment,
		IsAnonymous:     newReview.IsAnonymous,
		ReviewCreatedAt: newReview.ReviewCreatedAt,
	}

	rs.logger.Info("Review created successfully", zap.String("reviewID", response.ReviewID))

	return response, nil
}

// Helper function để update expert rating
func (rs *reviewService) updateExpertRating(tx *gorm.DB, expertProfileID uuid.UUID) error {
	var avgRating float64
	var totalReviews int64

	// Calculate new average rating
	err := tx.Table("tbl_consultation_reviews").
		Select("AVG(rating_score), COUNT(*)").
		Where("expert_profile_id = ? AND is_visible = true", expertProfileID).
		Row().Scan(&avgRating, &totalReviews)

	if err != nil {
		return fmt.Errorf("failed to calculate average rating: %w", err)
	}

	// Update expert profile
	return tx.Table("tbl_expert_profiles").
		Where("expert_profile_id = ?", expertProfileID).
		Updates(map[string]interface{}{
			"average_rating":    avgRating,
			"total_reviews":     totalReviews,
			"expert_updated_at": time.Now(),
		}).Error
}

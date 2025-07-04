package consultationreview

import (
	"cbs_backend/internal/modules/consultation_review/dtoreviews"
	"context"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	iReviewService IReviews
)

type IReviews interface {
	CreateReview(ctx context.Context, req dtoreviews.CreateReviewRequest) (*dtoreviews.CreateReviewResponse, error)
}

func InitReviewService(db *gorm.DB, logger zap.Logger) {
	iReviewService = NewReviewService(db, &logger)
}

func Review() IReviews {
	if iReviewService == nil {
		panic("ReviewService not initialized. Call InitReviewService(db,Logger) first.")
	}
	return iReviewService
}

package experts

import (
	dtoexperts "cbs_backend/internal/modules/experts/expertsdto"
	"cbs_backend/utils/cache"
	"context"

	"github.com/google/uuid"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	iExpertsService IExperts
)

func InitExpertService(db *gorm.DB, cache cache.ExpertCache, logger *zap.Logger) {
	iExpertsService = NewExpertService(db, cache, logger)
}
func Expert() IExperts {
	if iExpertsService == nil {
		panic("AuthService not initialized. Call InitAuthService(db) first.")
	}
	return iExpertsService
}

type IExperts interface {
	GetAllsExpert(ctx context.Context) (*[]dtoexperts.GetAllExpertsRespone, error)
	CreateExpertProfile(ctx context.Context, res dtoexperts.CreateProfileExpertRequest) (*dtoexperts.CreateProfileExpertResponse, error)
	UpdateExpertProfile(ctx context.Context, res dtoexperts.UpdateProfileExpertRequest) (*dtoexperts.UpdateProfileExpertResponse, error)
	GetExpertProfileDetails(ctx context.Context, expertid uuid.UUID) (*dtoexperts.ExpertFullDetailResponse, error)
	//Delete
	//Working hour
	CreateWorkHour(ctx context.Context, req dtoexperts.CreateWorkingHourRequest) (*dtoexperts.CreateWorkingHourResponse, error)
	UpdateWorkHour(ctx context.Context, req dtoexperts.UpdateWorkingHourRequest) (*dtoexperts.UpdateWorkingHourResponse, error)

	//Unvailable
	CreateUnavailableTime(ctx context.Context, req dtoexperts.CreateUnavailableTimeRequest) (*dtoexperts.CreateUnavailableTimeResponse, error)
	UpdateUnavailableTime(ctx context.Context, req dtoexperts.UpdateUnavailableTimeRequest) (*dtoexperts.UpdateUnavailableTimeResponse, error)
}

package experts

import (
	dtoexperts "cbs_backend/internal/modules/experts/expertsdto"
	"cbs_backend/utils/cache"
	"context"

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
	GetExpertProfileDetails(ctx context.Context, expertid string) (*dtoexperts.ExpertFullDetailResponse, error)
	//Delete

	//Working hour
	CreateWorkHour(ctx context.Context, req dtoexperts.CreateWorkingHourRequest) (*dtoexperts.CreateWorkingHourResponse, error)
	UpdateWorkHour(ctx context.Context, req dtoexperts.UpdateWorkingHourRequest) (*dtoexperts.UpdateWorkingHourResponse, error)
	GetAllWorkHourByExpertID(ctx context.Context, expertID string) ([]*dtoexperts.GetAllWorkingHourResponse, error)

	//Unavailable Time
	CreateUnavailableTime(ctx context.Context, req dtoexperts.CreateUnavailableTimeRequest) (*dtoexperts.CreateUnavailableTimeResponse, error)
	UpdateUnavailableTime(ctx context.Context, req dtoexperts.UpdateUnavailableTimeRequest) (*dtoexperts.UpdateUnavailableTimeResponse, error)
	GetAllUnavailableTimeByExpertID(ctx context.Context, expertID string) ([]*dtoexperts.GetAllsExpertUnavailableTimeResponse, error)

	//Expert Specializations
	CreateExpertSpecialization(ctx context.Context, req dtoexperts.CreateSpecializationRequest) (*dtoexperts.CreateSpecializationResponse, error)
	UpdateExpertSpecialization(ctx context.Context, req dtoexperts.UpdateExpertSpecializationRequest) (*dtoexperts.UpdateExpertSpecializationRespone, error)
	GetAllExpertSpecializationByExpertID(ctx context.Context, expertID string) ([]*dtoexperts.GetAllExpertSpecializationRespone, error)

	//PricingConfig
	CreatePrice(ctx context.Context, req dtoexperts.CreatePricingConfigRequest) (*dtoexperts.CreatePricingConfigResponse, error)
	UpdatePrice(ctx context.Context, req dtoexperts.UpdatePricingConfigRequest) (*dtoexperts.UpdatePricingConfigResponse, error)
	GetAllPriceByExpertID(ctx context.Context, expertID string) ([]*dtoexperts.PricingConfigResponse, error)

	DeleteExpertProfile(ctx context.Context, expertID string) error
	DeleteWorkHour(ctx context.Context, workingHourID string) error
	DeleteUnavailableTime(ctx context.Context, unavailableTimeID string) error
	DeleteExpertSpecialization(ctx context.Context, specializationID string) error
	DeletePrice(ctx context.Context, pricingID string) error
}

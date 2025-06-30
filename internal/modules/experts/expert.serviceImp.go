package experts

import (
	"cbs_backend/internal/common"
	entityexpert "cbs_backend/internal/modules/experts/entity"
	dtoexperts "cbs_backend/internal/modules/experts/expertsdto"
	"cbs_backend/internal/modules/users/entity"
	"cbs_backend/utils/cache"
	utils "cbs_backend/utils/cache"
	"context"

	"errors"
	"fmt"

	"github.com/google/uuid"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type expertService struct {
	db     *gorm.DB
	cache  utils.ExpertCache
	logger *zap.Logger
}

func NewExpertService(db *gorm.DB, cache cache.ExpertCache, logger *zap.Logger) *expertService {
	return &expertService{db: db, cache: cache, logger: logger}
}

func (es *expertService) CreateExpertProfile(ctx context.Context, req dtoexperts.CreateProfileExpertRequest) (*dtoexperts.CreateProfileExpertResponse, error) {
	// 1. Parse UserID from string to UUID
	userUUID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	// 2. Validate user exists and is active
	var user entity.User
	if err := es.db.WithContext(ctx).
		Where("user_id = ? AND is_active = true", userUUID).
		First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user with ID %s not found or inactive", req.UserID)
		}
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	// 3. Check if ExpertProfile already exists
	var existing entityexpert.ExpertProfile
	if err := es.db.WithContext(ctx).
		Where("user_id = ?", userUUID).
		First(&existing).Error; err == nil {
		return nil, fmt.Errorf("expert profile already exists for user %s", req.UserID)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check expert profile: %w", err)
	}

	// 4. Tạo ExpertProfile
	newProfile := entityexpert.ExpertProfile{
		UserID:             userUUID,
		SpecializationList: req.SpecializationList,
		ExperienceYears:    req.ExperienceYears,
		ExpertBio:          req.ExpertBio,
		ConsultationFee:    req.ConsultationFee,
		IsVerified:         false,
		LicenseNumber:      req.LicenseNumber,
		AvailableOnline:    req.AvailableOnline,
		AvailableOffline:   req.AvailableOffline,
	}

	if err := es.db.WithContext(ctx).Create(&newProfile).Error; err != nil {
		return nil, fmt.Errorf("failed to create expert profile: %w", err)
	}

	// 5. Preload User lại sau khi tạo (nếu cần dùng)
	if err := es.db.WithContext(ctx).
		Preload("User").
		First(&newProfile, "expert_profile_id = ?", newProfile.ExpertProfileID).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch expert with user: %w", err)
	}

	// 6. Kiểm tra User có được preload
	if newProfile.UserID == uuid.Nil {
		return nil, fmt.Errorf("user data not loaded in expert profile")
	}

	// 7. Map sang DTO
	userDTO := dtoexperts.UserDTO{
		UserID:    user.UserID.String(),
		FullName:  user.FullName,
		Email:     user.UserEmail,
		AvatarURL: user.AvatarURL,
	}

	resp := &dtoexperts.CreateProfileExpertResponse{
		ExpertProfileID:    newProfile.ExpertProfileID.String(),
		SpecializationList: newProfile.SpecializationList,
		ExperienceYears:    newProfile.ExperienceYears,
		ExpertBio:          newProfile.ExpertBio,
		ConsultationFee:    newProfile.ConsultationFee,
		AverageRating:      newProfile.AverageRating,
		TotalReviews:       newProfile.TotalReviews,
		IsVerified:         newProfile.IsVerified,
		LicenseNumber:      newProfile.LicenseNumber,
		AvailableOnline:    newProfile.AvailableOnline,
		AvailableOffline:   newProfile.AvailableOffline,
		User:               userDTO,
	}
	return resp, nil
}

func (es *expertService) GetExpertProfileDetails(ctx context.Context, expertID string) (*dtoexperts.ExpertFullDetailResponse, error) {
	// 1. Thử lấy từ cache
	cachedExpert, err := es.cache.GetExpertDetail(ctx, expertID)
	if err == nil && cachedExpert != nil {
		es.logger.Info("Expert detail loaded from Redis cache", zap.String("expertID", expertID))
		return cachedExpert, nil
	}
	// 2. Nếu không có trong cache => Load từ DB
	var expert entityexpert.ExpertProfile
	if err := es.db.WithContext(ctx).
		Preload("User").
		First(&expert, "expert_profile_id = ? AND is_verified = true", expertID).Error; err != nil {
		return nil, fmt.Errorf("cannot find expert profile: %w", err)
	}

	// Load giờ làm việc
	var workingHours []entityexpert.ExpertWorkingHour
	if err := es.db.WithContext(ctx).
		Where("expert_profile_id = ?", expertID).
		Find(&workingHours).Error; err != nil {
		return nil, fmt.Errorf("cannot fetch working hours: %w", err)
	}

	// Load thời gian không rảnh
	var unavailableTimes []entityexpert.ExpertUnavailableTime
	if err := es.db.WithContext(ctx).
		Where("expert_profile_id = ?", expertID).
		Find(&unavailableTimes).Error; err != nil {
		return nil, fmt.Errorf("cannot fetch unavailable times: %w", err)
	}

	var userDTO dtoexperts.UserDTO
	if expert.User != nil {
		userDTO = dtoexperts.UserDTO{
			UserID:    expert.User.UserID.String(),
			FullName:  expert.User.FullName,
			Email:     expert.User.UserEmail,
			AvatarURL: expert.User.AvatarURL,
		}
	} else {
		return nil, fmt.Errorf("user info not found for expert: %s", expertID)
	}

	// Map working hours
	var whDTOs []dtoexperts.WorkingHourDTO
	for _, wh := range workingHours {
		whDTOs = append(whDTOs, dtoexperts.WorkingHourDTO{
			DayOfWeek: fmt.Sprintf("%d", wh.DayOfWeek),
			StartTime: wh.StartTime.Format("15:04"), // or use wh.StartTime.String() if you want full time
			EndTime:   wh.EndTime.Format("15:04"),   // or use wh.EndTime.String()
		})
	}

	// Map unavailable times
	var uaDTOs []dtoexperts.UnavailableTimeDTO
	for _, ua := range unavailableTimes {
		uaDTOs = append(uaDTOs, dtoexperts.UnavailableTimeDTO{
			StartTime: ua.UnavailableStartDatetime,
			EndTime:   ua.UnavailableEndDatetime,
		})
	}

	// Response DTO
	expertFromDB := &dtoexperts.ExpertFullDetailResponse{
		ExpertProfileID:    expert.ExpertProfileID.String(),
		SpecializationList: expert.SpecializationList,
		ExperienceYears:    *expert.ExperienceYears,
		ExpertBio:          *expert.ExpertBio,
		ConsultationFee:    *expert.ConsultationFee,
		AverageRating:      expert.AverageRating,
		TotalReviews:       expert.TotalReviews,
		IsVerified:         expert.IsVerified,
		LicenseNumber:      *expert.LicenseNumber,
		AvailableOnline:    expert.AvailableOnline,
		AvailableOffline:   expert.AvailableOffline,
		User:               userDTO,
		WorkingHours:       whDTOs,
		UnavailableTimes:   uaDTOs,
	}
	// if err :=  utils.SetExpertDetail(ctx, es.redisCache, expertID.String(), expertFromDB); err != nil {
	// 	es.logger.Warn("Failed to cache expert detail", zap.Error(err))
	// }

	return expertFromDB, nil
}

func (es *expertService) UpdateExpertProfile(ctx context.Context, req dtoexperts.UpdateProfileExpertRequest) (*dtoexperts.UpdateProfileExpertResponse, error) {
	// Parse ExpertProfileID from string to UUID
	expertUUID, err := uuid.Parse(req.ExpertProfileID)
	if err != nil {
		return nil, fmt.Errorf("invalid expert profile ID format: %w", err)
	}

	var expert entityexpert.ExpertProfile

	// 1. Kiểm tra tồn tại ExpertProfile
	if err := es.db.WithContext(ctx).
		Preload("User").
		First(&expert, "expert_profile_id = ?", expertUUID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("expert profile not found")
		}
		return nil, fmt.Errorf("failed to load expert profile: %w", err)
	}

	// 2. Cập nhật thông tin
	expert.SpecializationList = req.SpecializationList
	expert.ExperienceYears = req.ExperienceYears
	expert.ExpertBio = req.ExpertBio
	expert.ConsultationFee = req.ConsultationFee
	expert.LicenseNumber = req.LicenseNumber
	expert.AvailableOnline = req.AvailableOnline
	expert.AvailableOffline = req.AvailableOffline

	// 3. Cập nhật vào DB
	if err := es.db.WithContext(ctx).Save(&expert).Error; err != nil {
		return nil, fmt.Errorf("failed to update expert profile: %w", err)
	}

	// 4. Map DTO để trả về
	userDTO := dtoexperts.UserDTO{
		UserID:    expert.User.UserID.String(),
		FullName:  expert.User.FullName,
		Email:     expert.User.UserEmail,
		AvatarURL: expert.User.AvatarURL,
	}

	return &dtoexperts.UpdateProfileExpertResponse{
		ExpertProfileID:    expert.ExpertProfileID.String(),
		SpecializationList: expert.SpecializationList,
		ExperienceYears:    expert.ExperienceYears,
		ExpertBio:          expert.ExpertBio,
		ConsultationFee:    expert.ConsultationFee,
		AverageRating:      expert.AverageRating,
		TotalReviews:       expert.TotalReviews,
		IsVerified:         expert.IsVerified,
		LicenseNumber:      expert.LicenseNumber,
		AvailableOnline:    expert.AvailableOnline,
		AvailableOffline:   expert.AvailableOffline,
		User:               userDTO,
	}, nil
}

func (es *expertService) GetAllsExpert(ctx context.Context) (*[]dtoexperts.GetAllExpertsRespone, error) {
	var experts []entityexpert.ExpertProfile

	// 1. Lấy tất cả expert + preload user (fixed the WHERE clause position)
	if err := es.db.WithContext(ctx).
		Preload("User").
		Where("is_verified = true").
		Find(&experts).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch experts: %w", err)
	}

	// 2. Map sang DTO
	var result []dtoexperts.GetAllExpertsRespone
	for _, expert := range experts {
		var userDTO dtoexperts.UserDTO
		if expert.User != nil {
			userDTO = dtoexperts.UserDTO{
				UserID:    expert.User.UserID.String(),
				FullName:  expert.User.FullName,
				Email:     expert.User.UserEmail,
				AvatarURL: expert.User.AvatarURL,
			}
		}

		dto := dtoexperts.GetAllExpertsRespone{
			ExpertProfileID:    expert.ExpertProfileID.String(),
			SpecializationList: expert.SpecializationList,
			ExperienceYears:    expert.ExperienceYears,
			ConsultationFee:    expert.ConsultationFee,
			AverageRating:      expert.AverageRating,
			TotalReviews:       expert.TotalReviews,
			User:               userDTO,
		}
		result = append(result, dto)
	}

	return &result, nil
}

// Working hour
func (es *expertService) CreateWorkHour(ctx context.Context, req dtoexperts.CreateWorkingHourRequest) (*dtoexperts.CreateWorkingHourResponse, error) {
	// Parse ExpertProfileID from string to UUID
	expertUUID, err := uuid.Parse(req.ExpertProfileID)
	if err != nil {
		return nil, fmt.Errorf("invalid expert profile ID format: %w", err)
	}

	workingHour := entityexpert.ExpertWorkingHour{
		ExpertProfileID: expertUUID,
		DayOfWeek:       req.DayOfWeek,
		StartTime:       req.StartTime,
		EndTime:         req.EndTime,
		IsActive:        true,
	}

	if err := es.db.WithContext(ctx).Create(&workingHour).Error; err != nil {
		return nil, fmt.Errorf("insert work hour failed: %w", err)
	}

	return &dtoexperts.CreateWorkingHourResponse{
		WorkingHourID: workingHour.WorkingHourID.String(),
		DayOfWeek:     workingHour.DayOfWeek,
		StartTime:     workingHour.StartTime,
		EndTime:       workingHour.EndTime,
	}, nil
}

func (es *expertService) UpdateWorkHour(ctx context.Context, req dtoexperts.UpdateWorkingHourRequest) (*dtoexperts.UpdateWorkingHourResponse, error) {
	var wh entityexpert.ExpertWorkingHour
	if err := es.db.WithContext(ctx).First(&wh, "working_hour_id = ?", req.WorkingHourID).Error; err != nil {
		return nil, fmt.Errorf("work hour not found: %w", err)
	}

	wh.DayOfWeek = req.DayOfWeek
	wh.StartTime = req.StartTime
	wh.EndTime = req.EndTime

	if err := es.db.WithContext(ctx).Save(&wh).Error; err != nil {
		return nil, fmt.Errorf("update work hour failed: %w", err)
	}

	return &dtoexperts.UpdateWorkingHourResponse{
		WorkingHourID: wh.WorkingHourID.String(),
		DayOfWeek:     wh.DayOfWeek,
		StartTime:     wh.StartTime,
		EndTime:       wh.EndTime,
		IsActive:      wh.IsActive,
	}, nil
}
func (es *expertService) GetAllWorkHourByExpertID(ctx context.Context, expertID string) ([]*dtoexperts.GetAllWorkingHourResponse, error) {

	if expertID == "" {
		return nil, fmt.Errorf("expert ID must not be empty")
	}

	// ✅ slice đúng kiểu (con trỏ ‑ hoặc bỏ * tùy DTO bạn định trả)
	var workingHours []*dtoexperts.GetAllWorkingHourResponse

	// Ví dụ dùng GORM: bảng working_hours có cột expert_id
	if err := es.db.WithContext(ctx).
		Table("tbl_working_hours").
		Where("expert_id = ?", expertID).
		Order("day_of_week, start_time").
		Find(&workingHours).Error; err != nil {

		return nil, fmt.Errorf("query working hours: %w", err)
	}

	return workingHours, nil
}

// Unvailable Time
func (es *expertService) CreateUnavailableTime(ctx context.Context, req dtoexperts.CreateUnavailableTimeRequest) (*dtoexperts.CreateUnavailableTimeResponse, error) {
	// Parse ExpertProfileID from string to UUID
	expertUUID, err := uuid.Parse(req.ExpertProfileID)
	if err != nil {
		return nil, fmt.Errorf("invalid expert profile ID format: %w", err)
	}

	var recurrencePattern common.JSONB
	if req.RecurrencePattern != nil {
		recurrencePattern, _ = req.RecurrencePattern.(common.JSONB)
	}
	unavailable := entityexpert.ExpertUnavailableTime{
		ExpertProfileID:          expertUUID,
		UnavailableStartDatetime: req.StartDatetime,
		UnavailableEndDatetime:   req.EndDatetime,
		UnavailableReason:        req.Reason,
		IsRecurring:              req.IsRecurring,
		RecurrencePattern:        recurrencePattern,
	}

	if err := es.db.WithContext(ctx).Create(&unavailable).Error; err != nil {
		return nil, fmt.Errorf("create unavailable time failed: %w", err)
	}

	return &dtoexperts.CreateUnavailableTimeResponse{
		UnavailableTimeID: unavailable.UnavailableTimeID.String(),
		StartTime:         unavailable.UnavailableStartDatetime,
		EndTime:           unavailable.UnavailableEndDatetime,
		Reason:            unavailable.UnavailableReason,
		IsRecurring:       unavailable.IsRecurring,
	}, nil
}
func (es *expertService) UpdateUnavailableTime(ctx context.Context, req dtoexperts.UpdateUnavailableTimeRequest) (*dtoexperts.UpdateUnavailableTimeResponse, error) {
	var ua entityexpert.ExpertUnavailableTime
	if err := es.db.WithContext(ctx).First(&ua, "unavailable_time_id = ?", req.UnavailableTimeID).Error; err != nil {
		return nil, fmt.Errorf("unavailable time not found: %w", err)
	}

	ua.UnavailableStartDatetime = req.StartDatetime
	ua.UnavailableEndDatetime = req.EndDatetime
	ua.UnavailableReason = req.Reason
	ua.IsRecurring = req.IsRecurring
	if req.RecurrencePattern != nil {
		if rp, ok := req.RecurrencePattern.(common.JSONB); ok {
			ua.RecurrencePattern = rp
		} else {
			return nil, fmt.Errorf("invalid recurrence pattern type")
		}
	} else {
		ua.RecurrencePattern = nil
	}

	if err := es.db.WithContext(ctx).Save(&ua).Error; err != nil {
		return nil, fmt.Errorf("update unavailable time failed: %w", err)
	}

	return &dtoexperts.UpdateUnavailableTimeResponse{
		UnavailableTimeID: ua.UnavailableTimeID.String(),
		StartTime:         ua.UnavailableStartDatetime,
		EndTime:           ua.UnavailableEndDatetime,
		Reason:            ua.UnavailableReason,
		IsRecurring:       ua.IsRecurring,
	}, nil
}
func (es *expertService) GetAllUnavailableTimeByExpertID(ctx context.Context, expertID string) ([]*dtoexperts.GetAllsExpertUnavailableTimeResponse, error) {
	if expertID == "" {
		return nil, fmt.Errorf("expert Id must be not empty")
	}
	var result []*dtoexperts.GetAllsExpertUnavailableTimeResponse

	if err := es.db.WithContext(ctx).
		Table("tbl_expert_unavailable_times").
		Where("expert_profile_id = ?", expertID).
		Order("unavailable_start_datetime").
		Select(`
        unavailable_time_id        AS unavailable_time_id,
        unavailable_start_datetime AS start_time,
        unavailable_end_datetime   AS end_time,
        unavailable_reason         AS reason,
        is_recurring,
        recurrence_pattern
    `).
		Scan(&result).Error; err != nil {
		return nil, fmt.Errorf("query is failed %w", err)
	}

	return result, nil
}

// Expert Specialization
func (es *expertService) CreateExpertSpecialization(ctx context.Context, req dtoexperts.CreateSpecializationRequest) (*dtoexperts.CreateSpecializationResponse, error) {
	specialization := entityexpert.ExpertSpecialization{
		ExpertProfileID:           uuid.MustParse(req.ExpertProfileID),
		SpecializationName:        req.SpecializationName,
		SpecializationDescription: &req.SpecializationDescription,
		IsPrimary:                 req.IsPrimary,
	}

	if err := es.db.WithContext(ctx).Create(&specialization).Error; err != nil {
		return nil, fmt.Errorf("failed to create specialization: %w", err)
	}

	return &dtoexperts.CreateSpecializationResponse{
		SpecializationID:          specialization.SpecializationID.String(),
		ExpertProfileID:           specialization.ExpertProfileID.String(),
		SpecializationName:        specialization.SpecializationName,
		SpecializationDescription: *specialization.SpecializationDescription,
		IsPrimary:                 specialization.IsPrimary,
		CreateAt:                  specialization.CreatedAt,
	}, nil
}
func (es *expertService) UpdateExpertSpecialization(ctx context.Context, req dtoexperts.UpdateExpertSpecializationRequest) (*dtoexperts.UpdateExpertSpecializationRespone, error) {
	var specialization entityexpert.ExpertSpecialization

	if err := es.db.WithContext(ctx).
		First(&specialization, "specialization_id = ?", req.SpecializationID).Error; err != nil {
		return nil, fmt.Errorf("specialization not found: %w", err)
	}

	specialization.SpecializationName = req.SpecializationName
	specialization.SpecializationDescription = &req.SpecializationDescription
	specialization.IsPrimary = req.IsPrimary

	if err := es.db.WithContext(ctx).Save(&specialization).Error; err != nil {
		return nil, fmt.Errorf("failed to update specialization: %w", err)
	}

	return &dtoexperts.UpdateExpertSpecializationRespone{
		SpecializationID:          specialization.SpecializationID.String(),
		ExpertProfileID:           specialization.ExpertProfileID.String(),
		SpecializationName:        specialization.SpecializationName,
		SpecializationDescription: *specialization.SpecializationDescription,
		IsPrimary:                 specialization.IsPrimary,
	}, nil
}

func (es *expertService) GetAllExpertSpecializationByExpertID(ctx context.Context, expertID string) ([]*dtoexperts.GetAllExpertSpecializationRespone, error) {
	// Chuyển expertID sang UUID nếu cần (nếu entity dùng uuid.UUID)
	expertUUID, err := uuid.Parse(expertID)
	if err != nil {
		return nil, fmt.Errorf("invalid expert ID: %w", err)
	}

	// Query DB
	var specializations []entityexpert.ExpertSpecialization
	if err := es.db.WithContext(ctx).
		Where("expert_profile_id = ?", expertUUID).
		Order("is_primary DESC, created_at ASC").
		Find(&specializations).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve specializations: %w", err)
	}

	// Map từ entity sang DTO
	var responses []*dtoexperts.GetAllExpertSpecializationRespone
	for _, s := range specializations {
		resp := &dtoexperts.GetAllExpertSpecializationRespone{
			SpecializationID:          s.SpecializationID.String(),
			ExpertProfileID:           s.ExpertProfileID.String(),
			SpecializationName:        s.SpecializationName,
			SpecializationDescription: *s.SpecializationDescription,
			IsPrimary:                 s.IsPrimary,
			CreateAt:                  s.CreatedAt,
		}
		responses = append(responses, resp)
	}

	return responses, nil
}

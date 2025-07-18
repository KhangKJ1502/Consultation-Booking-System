package experts

import (
	"cbs_backend/internal/common"
	entityexpert "cbs_backend/internal/modules/experts/entity"
	dtoexperts "cbs_backend/internal/modules/experts/expertsdto"
	"cbs_backend/internal/modules/users/entity"
	"cbs_backend/utils/cache"
	utils "cbs_backend/utils/cache"
	"context"
	"time"

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
		startParsed, _ := time.Parse("15:04:05", wh.StartTime)
		endParsed, _ := time.Parse("15:04:05", wh.EndTime)

		whDTOs = append(whDTOs, dtoexperts.WorkingHourDTO{
			DayOfWeek: fmt.Sprintf("%d", wh.DayOfWeek),
			StartTime: startParsed.Format("15:04"),
			EndTime:   endParsed.Format("15:04"),
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
	// Load pricing configs
	var pricings []entityexpert.PricingConfig
	if err := es.db.WithContext(ctx).Where("expert_profile_id = ?", expertID).Find(&pricings).Error; err != nil {
		return nil, fmt.Errorf("cannot fetch pricing configs: %w", err)
	}
	var pricingDTOs []dtoexperts.PricingConfigResponse
	for _, p := range pricings {
		pricingDTOs = append(pricingDTOs, dtoexperts.PricingConfigResponse{
			PricingID:          p.PricingID.String(),
			ServiceType:        p.ServiceType,
			ConsultationType:   p.ConsultationType,
			DurationMinutes:    p.DurationMinutes,
			BasePrice:          p.BasePrice,
			DiscountPercentage: p.DiscountPercentage,
			IsActive:           p.IsActive,
			ValidFrom:          p.ValidFrom,
			ValidUntil:         p.ValidUntil,
			PricingCreatedAt:   p.PricingCreatedAt,
		})
	}

	// Load specializations
	var specs []entityexpert.ExpertSpecialization
	if err := es.db.WithContext(ctx).Where("expert_profile_id = ?", expertID).Find(&specs).Error; err != nil {
		return nil, fmt.Errorf("cannot fetch specializations: %w", err)
	}
	var specDTOs []dtoexperts.GetAllExpertSpecializationRespone
	for _, s := range specs {
		specDTOs = append(specDTOs, dtoexperts.GetAllExpertSpecializationRespone{
			SpecializationID:          s.SpecializationID.String(),
			ExpertProfileID:           s.ExpertProfileID.String(),
			SpecializationName:        s.SpecializationName,
			SpecializationDescription: *s.SpecializationDescription,
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
	if err := es.cache.SetExpertDetail(ctx, expertID, expertFromDB); err != nil {
		es.logger.Warn("Failed to cache expert detail", zap.Error(err))
	}
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

	// 5. Cache operations - use correct variable name
	_ = es.cache.DeleteExpertDetail(ctx, expertUUID.String()) // Fixed: use expertUUID instead of undefined expertID

	// Cache the updated expert detail
	expertDetail := &dtoexperts.ExpertFullDetailResponse{
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
	}

	// Set the updated data in cache
	if err := es.cache.SetExpertDetail(ctx, expertUUID.String(), expertDetail); err != nil {
		es.logger.Warn("Failed to update expert cache", zap.Error(err))
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

func (es *expertService) DeleteExpertProfile(ctx context.Context, expertID string) error {
	expertUUID, err := uuid.Parse(expertID)
	if err != nil {
		return fmt.Errorf("invalid expert ID format: %w", err)
	}

	// Soft delete or hard delete depending on your business logic
	if err := es.db.WithContext(ctx).Delete(&entityexpert.ExpertProfile{}, "expert_profile_id = ?", expertUUID).Error; err != nil {
		return fmt.Errorf("failed to delete expert profile: %w", err)
	}

	// Clear cache
	_ = es.cache.DeleteExpertDetail(ctx, expertID)

	return nil
}

// Working hour
func (es *expertService) CreateWorkHour(ctx context.Context, req dtoexperts.CreateWorkingHourRequest) (*dtoexperts.CreateWorkingHourResponse, error) {
	expertUUID, err := uuid.Parse(req.ExpertProfileID)
	if err != nil {
		return nil, fmt.Errorf("invalid expert profile ID format: %w", err)
	}

	// Parse chuỗi thành time.Time
	startTimeParsed, err := time.Parse("15:04", req.StartTime)
	if err != nil {
		return nil, fmt.Errorf("invalid start_time format: %w", err)
	}

	endTimeParsed, err := time.Parse("15:04", req.EndTime)
	if err != nil {
		return nil, fmt.Errorf("invalid end_time format: %w", err)
	}

	workingHour := entityexpert.ExpertWorkingHour{
		ExpertProfileID: expertUUID,
		DayOfWeek:       req.DayOfWeek,
		StartTime:       startTimeParsed.Format("15:04:05"),
		EndTime:         endTimeParsed.Format("15:04:05"),
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
		IsActive:      workingHour.IsActive,
	}, nil
}

func (es *expertService) UpdateWorkHour(ctx context.Context, req dtoexperts.UpdateWorkingHourRequest) (*dtoexperts.UpdateWorkingHourResponse, error) {
	var wh entityexpert.ExpertWorkingHour
	if err := es.db.WithContext(ctx).First(&wh, "working_hour_id = ?", req.WorkingHourID).Error; err != nil {
		return nil, fmt.Errorf("work hour not found: %w", err)
	}

	// Parse chuỗi thời gian từ request
	startTimeParsed, err := time.Parse("15:04", req.StartTime)
	if err != nil {
		return nil, fmt.Errorf("invalid start_time format: %w", err)
	}

	endTimeParsed, err := time.Parse("15:04", req.EndTime)
	if err != nil {
		return nil, fmt.Errorf("invalid end_time format: %w", err)
	}

	// Gán vào entity (sẽ lưu dưới dạng "08:00:00")
	wh.DayOfWeek = req.DayOfWeek
	wh.StartTime = startTimeParsed.Format("15:04:05")
	wh.EndTime = endTimeParsed.Format("15:04:05")

	if err := es.db.WithContext(ctx).Save(&wh).Error; err != nil {
		return nil, fmt.Errorf("update work hour failed: %w", err)
	}

	// Parse lại từ chuỗi để trả response
	startTimeReturn, _ := time.Parse("15:04:05", wh.StartTime)
	endTimeReturn, _ := time.Parse("15:04:05", wh.EndTime)

	return &dtoexperts.UpdateWorkingHourResponse{
		WorkingHourID: wh.WorkingHourID.String(),
		DayOfWeek:     wh.DayOfWeek,
		StartTime:     startTimeReturn,
		EndTime:       endTimeReturn,
		IsActive:      wh.IsActive,
	}, nil
}

func (es *expertService) GetAllWorkHourByExpertID(ctx context.Context, expertID string) ([]*dtoexperts.GetAllWorkingHourResponse, error) {

	if expertID == "" {
		return nil, fmt.Errorf("expert ID must not be empty")
	}

	// ✅ slice đúng kiểu (con trỏ ‑ hoặc bỏ * tùy DTO bạn định trả)
	var workingHours []*dtoexperts.GetAllWorkingHourResponse

	if err := es.db.WithContext(ctx).
		Table("tbl_expert_working_hours").
		Where("expert_profile_id = ?", expertID).
		Order("day_of_week, start_time").
		Find(&workingHours).Error; err != nil {

		return nil, fmt.Errorf("query working hours: %w", err)
	}

	return workingHours, nil
}

func (es *expertService) DeleteWorkHour(ctx context.Context, workingHourID string) error {
	workingHourUUID, err := uuid.Parse(workingHourID)
	if err != nil {
		return fmt.Errorf("invalid working hour ID format: %w", err)
	}

	if err := es.db.WithContext(ctx).Delete(&entityexpert.ExpertWorkingHour{}, "working_hour_id = ?", workingHourUUID).Error; err != nil {
		return fmt.Errorf("failed to delete working hour: %w", err)
	}

	return nil
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

	// Kiểm tra xem bản ghi có tồn tại không
	if err := es.db.WithContext(ctx).First(&ua, "unavailable_time_id = ?", req.UnavailableTimeID).Error; err != nil {
		return nil, fmt.Errorf("unavailable time not found: %w", err)
	}

	// Parse thời gian
	startTime, err := time.Parse(time.RFC3339, req.StartDatetime)
	if err != nil {
		return nil, fmt.Errorf("invalid start_datetime format (expecting RFC3339): %w", err)
	}

	endTime, err := time.Parse(time.RFC3339, req.EndDatetime)
	if err != nil {
		return nil, fmt.Errorf("invalid end_datetime format (expecting RFC3339): %w", err)
	}

	// Gán dữ liệu vào entity
	ua.UnavailableStartDatetime = startTime
	ua.UnavailableEndDatetime = endTime
	ua.UnavailableReason = req.Reason
	ua.IsRecurring = req.IsRecurring

	// Xử lý recurrence pattern nếu có
	if req.RecurrencePattern != nil {
		if rp, ok := req.RecurrencePattern.(common.JSONB); ok {
			ua.RecurrencePattern = rp
		} else {
			return nil, fmt.Errorf("invalid recurrence pattern type, must be JSONB")
		}
	} else {
		ua.RecurrencePattern = nil
	}

	// Cập nhật DB
	if err := es.db.WithContext(ctx).Save(&ua).Error; err != nil {
		return nil, fmt.Errorf("update unavailable time failed: %w", err)
	}

	// Trả về response
	return &dtoexperts.UpdateUnavailableTimeResponse{
		UnavailableTimeID: ua.UnavailableTimeID.String(),
		StartTime:         ua.UnavailableStartDatetime,
		EndTime:           ua.UnavailableEndDatetime,
		Reason:            ua.UnavailableReason,
		IsRecurring:       ua.IsRecurring,
		RecurrencePattern: ua.RecurrencePattern,
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

func (es *expertService) DeleteUnavailableTime(ctx context.Context, unavailableTimeID string) error {
	unavailableTimeUUID, err := uuid.Parse(unavailableTimeID)
	if err != nil {
		return fmt.Errorf("invalid unavailable time ID format: %w", err)
	}

	if err := es.db.WithContext(ctx).Delete(&entityexpert.ExpertUnavailableTime{}, "unavailable_time_id = ?", unavailableTimeUUID).Error; err != nil {
		return fmt.Errorf("failed to delete unavailable time: %w", err)
	}

	return nil
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

func (es *expertService) DeleteExpertSpecialization(ctx context.Context, specializationID string) error {
	specializationUUID, err := uuid.Parse(specializationID)
	if err != nil {
		return fmt.Errorf("invalid specialization ID format: %w", err)
	}

	if err := es.db.WithContext(ctx).Delete(&entityexpert.ExpertSpecialization{}, "specialization_id = ?", specializationUUID).Error; err != nil {
		return fmt.Errorf("failed to delete specialization: %w", err)
	}

	return nil
}

// Pricing
func (es *expertService) CreatePrice(ctx context.Context, req dtoexperts.CreatePricingConfigRequest) (*dtoexperts.CreatePricingConfigResponse, error) {

	ExpertID, err := uuid.Parse(req.ExpertProfileID)
	if err != nil {
		return nil, fmt.Errorf("expert id is empty")
	}
	pricingConfig := &entityexpert.PricingConfig{
		ExpertProfileID:    &ExpertID,
		ServiceType:        req.ServiceType,
		ConsultationType:   req.ConsultationType,
		DurationMinutes:    req.DurationMinutes,
		BasePrice:          req.BasePrice,
		DiscountPercentage: req.DiscountPercentage,
		ValidFrom:          req.ValidFrom,
		ValidUntil:         req.ValidUntil,
	}
	if err := es.db.WithContext(ctx).Create(&pricingConfig).Error; err != nil {
		return nil, fmt.Errorf("failed to create Pricing Config: %w", err)
	}
	return &dtoexperts.CreatePricingConfigResponse{
		PricingID:          pricingConfig.PricingID.String(),
		ExpertProfileID:    pricingConfig.ExpertProfileID.String(),
		ServiceType:        pricingConfig.ServiceType,
		ConsultationType:   pricingConfig.ConsultationType,
		DurationMinutes:    pricingConfig.DurationMinutes,
		BasePrice:          pricingConfig.BasePrice,
		DiscountPercentage: pricingConfig.DiscountPercentage,
		IsActive:           pricingConfig.IsActive,
		ValidFrom:          pricingConfig.ValidFrom,
		ValidUntil:         pricingConfig.ValidUntil,
		PricingCreatedAt:   pricingConfig.PricingCreatedAt,
	}, nil
}
func (es *expertService) UpdatePrice(ctx context.Context, req dtoexperts.UpdatePricingConfigRequest) (*dtoexperts.UpdatePricingConfigResponse, error) {
	var pricingConfig entityexpert.PricingConfig

	// Tìm pricing theo ID
	if err := es.db.WithContext(ctx).First(&pricingConfig, "pricing_id = ?", req.PricingID).Error; err != nil {
		return nil, fmt.Errorf("pricing config not found: %w", err)
	}

	// Cập nhật các trường
	pricingConfig.ServiceType = req.ServiceType
	pricingConfig.ConsultationType = req.ConsultationType
	pricingConfig.DurationMinutes = req.DurationMinutes
	pricingConfig.BasePrice = req.BasePrice
	pricingConfig.DiscountPercentage = req.DiscountPercentage
	pricingConfig.ValidFrom = req.ValidFrom
	pricingConfig.ValidUntil = req.ValidUntil
	pricingConfig.IsActive = req.IsActive

	// Save vào DB
	if err := es.db.WithContext(ctx).Save(&pricingConfig).Error; err != nil {
		return nil, fmt.Errorf("failed to update pricing config: %w", err)
	}

	return &dtoexperts.UpdatePricingConfigResponse{
		PricingID: pricingConfig.PricingID.String(),
	}, nil
}

func (es *expertService) GetAllPriceByExpertID(ctx context.Context, expertID string) ([]*dtoexperts.PricingConfigResponse, error) {
	if expertID == "" {
		return nil, fmt.Errorf("expert Id must be not empty")
	}

	var pricingConfigs []entityexpert.PricingConfig
	// Tìm tất cả pricing theo ExpertProfileID
	if err := es.db.WithContext(ctx).
		Where("expert_profile_id = ?", expertID).
		Find(&pricingConfigs).Error; err != nil {
		return nil, fmt.Errorf("failed to get pricing configs: %w", err)
	}

	// Mapping sang DTO
	var res []*dtoexperts.PricingConfigResponse
	for _, p := range pricingConfigs {
		res = append(res, &dtoexperts.PricingConfigResponse{
			PricingID:          p.PricingID.String(),
			ExpertProfileID:    p.ExpertProfileID.String(),
			ServiceType:        p.ServiceType,
			ConsultationType:   p.ConsultationType,
			DurationMinutes:    p.DurationMinutes,
			BasePrice:          p.BasePrice,
			DiscountPercentage: p.DiscountPercentage,
			IsActive:           p.IsActive,
			ValidFrom:          p.ValidFrom,
			ValidUntil:         p.ValidUntil,
			PricingCreatedAt:   p.PricingCreatedAt,
		})
	}

	return res, nil
}

func (es *expertService) DeletePrice(ctx context.Context, pricingID string) error {
	pricingUUID, err := uuid.Parse(pricingID)
	if err != nil {
		return fmt.Errorf("invalid pricing ID format: %w", err)
	}

	if err := es.db.WithContext(ctx).Delete(&entityexpert.PricingConfig{}, "pricing_id = ?", pricingUUID).Error; err != nil {
		return fmt.Errorf("failed to delete pricing config: %w", err)
	}

	return nil
}

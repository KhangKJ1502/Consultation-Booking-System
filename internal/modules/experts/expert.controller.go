package experts

import (
	dtoexperts "cbs_backend/internal/modules/experts/expertsdto"
	"cbs_backend/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ExpertController struct{}

func NewExpertController() *ExpertController {
	return &ExpertController{}
}

func (ec *ExpertController) GetAllExpert(ctx *gin.Context) (res interface{}, err error) {
	req, err := Expert().GetAllsExpert(ctx)
	if err != nil {
		return "", response.NewAPIError(http.StatusBadRequest, "Invalid request payload", err.Error())
	}
	return req, nil
}

func (ec *ExpertController) CreateExpertProfile(ctx *gin.Context) (res interface{}, err error) {
	var req dtoexperts.CreateProfileExpertRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid request payload", err.Error())
	}
	resNewExpert, err := Expert().CreateExpertProfile(ctx, req)
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "create expert is failed", err.Error())
	}
	return resNewExpert, nil
}

func (ec *ExpertController) UpdateExpertProfile(ctx *gin.Context) (res interface{}, err error) {
	var req dtoexperts.UpdateProfileExpertRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid request payload", err.Error())
	}
	resUpdateExpert, err := Expert().UpdateExpertProfile(ctx, req)
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "update expert is failed", err.Error())
	}
	return resUpdateExpert, nil
}

func (ec *ExpertController) GetExpertProfileDetails(ctx *gin.Context) (res interface{}, err error) {
	idExpertStr := ctx.Param("id")
	expertID := idExpertStr
	resExpertDetail, err := Expert().GetExpertProfileDetails(ctx, expertID)
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid response load detail", err.Error())
	}
	return resExpertDetail, nil
}

// Working Hour Controllers
func (ec *ExpertController) CreateWorkHour(ctx *gin.Context) (res interface{}, err error) {
	var req dtoexperts.CreateWorkingHourRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid request payload", err.Error())
	}
	resNewWorkHour, err := Expert().CreateWorkHour(ctx, req)
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "create work hour is failed", err.Error())
	}
	return resNewWorkHour, nil
}

func (ec *ExpertController) UpdateWorkHour(ctx *gin.Context) (res interface{}, err error) {
	var req dtoexperts.UpdateWorkingHourRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid request payload", err.Error())
	}
	resUpdateWorkHour, err := Expert().UpdateWorkHour(ctx, req)
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "update work hour is failed", err.Error())
	}
	return resUpdateWorkHour, nil
}

func (ec *ExpertController) GetAllWorkHourByExpertID(ctx *gin.Context) (res interface{}, err error) {
	expertID := ctx.Param("expertId")
	if expertID == "" {
		return nil, response.NewAPIError(http.StatusBadRequest, "Expert ID is required", "Expert ID parameter is missing")
	}

	resWorkHours, err := Expert().GetAllWorkHourByExpertID(ctx, expertID)
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Failed to get work hours", err.Error())
	}
	return resWorkHours, nil
}

// Unavailable Time Controllers
func (ec *ExpertController) CreateUnavailableTime(ctx *gin.Context) (res interface{}, err error) {
	var req dtoexperts.CreateUnavailableTimeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid request payload", err.Error())
	}
	resNewUnavailableTime, err := Expert().CreateUnavailableTime(ctx, req)
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "create unavailable time is failed", err.Error())
	}
	return resNewUnavailableTime, nil
}

func (ec *ExpertController) UpdateUnavailableTime(ctx *gin.Context) (res interface{}, err error) {
	var req dtoexperts.UpdateUnavailableTimeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid request payload", err.Error())
	}
	resUpdateUnavailableTime, err := Expert().UpdateUnavailableTime(ctx, req)
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "update unavailable time is failed", err.Error())
	}
	return resUpdateUnavailableTime, nil
}

func (ec *ExpertController) GetAllUnavailableTimeByExpertID(ctx *gin.Context) (res interface{}, err error) {
	expertID := ctx.Param("expertId")
	if expertID == "" {
		return nil, response.NewAPIError(http.StatusBadRequest, "Expert ID is required", "Expert ID parameter is missing")
	}

	resUnavailableTimes, err := Expert().GetAllUnavailableTimeByExpertID(ctx, expertID)
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Failed to get unavailable times", err.Error())
	}
	return resUnavailableTimes, nil
}

// -------------------- Specialization Controllers --------------------

// Tạo mới specialization cho expert
func (ec *ExpertController) CreateExpertSpecialization(ctx *gin.Context) (res interface{}, err error) {
	var req dtoexperts.CreateSpecializationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid request payload", err.Error())
	}

	newSpec, err := Expert().CreateExpertSpecialization(ctx, req)
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "create specialization failed", err.Error())
	}
	return newSpec, nil
}

// Cập nhật specialization
func (ec *ExpertController) UpdateExpertSpecialization(ctx *gin.Context) (res interface{}, err error) {
	var req dtoexperts.UpdateExpertSpecializationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid request payload", err.Error())
	}

	updatedSpec, err := Expert().UpdateExpertSpecialization(ctx, req)
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "update specialization failed", err.Error())
	}
	return updatedSpec, nil
}

// Lấy danh sách specialization theo ExpertID
func (ec *ExpertController) GetAllExpertSpecializationByExpertID(ctx *gin.Context) (res interface{}, err error) {
	expertID := ctx.Param("expertId")
	if expertID == "" {
		return nil, response.NewAPIError(http.StatusBadRequest, "Expert ID is required", "Expert ID parameter is missing")
	}

	specs, err := Expert().GetAllExpertSpecializationByExpertID(ctx, expertID)
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "failed to get specializations", err.Error())
	}
	return specs, nil
}

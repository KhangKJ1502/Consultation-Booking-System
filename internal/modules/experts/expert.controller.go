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

//Expert

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
func (ec *ExpertController) DeleteExpertProfile(ctx *gin.Context) (res interface{}, err error) {
	expertID := ctx.Param("expertId")
	if expertID == "" {
		return nil, response.NewAPIError(http.StatusBadRequest, "Expert ID is required", "Expert ID parameter is missing")
	}

	err = Expert().DeleteExpertProfile(ctx, expertID)
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "delete expert profile failed", err.Error())
	}
	return map[string]string{"message": "Expert profile deleted successfully"}, nil
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

func (ec *ExpertController) DeleteWorkHour(ctx *gin.Context) (res interface{}, err error) {
	workingHourID := ctx.Param("workingHourId")
	if workingHourID == "" {
		return nil, response.NewAPIError(http.StatusBadRequest, "Working Hour ID is required", "Working Hour ID parameter is missing")
	}

	err = Expert().DeleteWorkHour(ctx, workingHourID)
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "delete work hour failed", err.Error())
	}
	return map[string]string{"message": "Work hour deleted successfully"}, nil
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
func (ec *ExpertController) DeleteUnavailableTime(ctx *gin.Context) (res interface{}, err error) {
	unavailableTimeID := ctx.Param("unavailableTimeId")
	if unavailableTimeID == "" {
		return nil, response.NewAPIError(http.StatusBadRequest, "Unavailable Time ID is required", "Unavailable Time ID parameter is missing")
	}

	err = Expert().DeleteUnavailableTime(ctx, unavailableTimeID)
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "delete unavailable time failed", err.Error())
	}
	return map[string]string{"message": "Unavailable time deleted successfully"}, nil
}

// -------------------- Specialization Controllers --------------------

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

func (ec *ExpertController) DeleteExpertSpecialization(ctx *gin.Context) (res interface{}, err error) {
	specializationID := ctx.Param("specializationId")
	if specializationID == "" {
		return nil, response.NewAPIError(http.StatusBadRequest, "Specialization ID is required", "Specialization ID parameter is missing")
	}

	err = Expert().DeleteExpertSpecialization(ctx, specializationID)
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "delete specialization failed", err.Error())
	}
	return map[string]string{"message": "Specialization deleted successfully"}, nil
}

// -----------Pricing Config Controller
func (ec *ExpertController) CreatePrice(ctx *gin.Context) (res interface{}, err error) {
	var req dtoexperts.CreatePricingConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid request payload", err.Error())
	}

	resNewPrice, err := Expert().CreatePrice(ctx, req)
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "create pricing config failed", err.Error())
	}
	return resNewPrice, nil
}
func (ec *ExpertController) UpdatePrice(ctx *gin.Context) (res interface{}, err error) {
	var req dtoexperts.UpdatePricingConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid request payload", err.Error())
	}

	resUpdatedPrice, err := Expert().UpdatePrice(ctx, req)
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "update pricing config failed", err.Error())
	}
	return resUpdatedPrice, nil
}
func (ec *ExpertController) GetAllPriceByExpertID(ctx *gin.Context) (res interface{}, err error) {
	expertID := ctx.Param("expertId")
	if expertID == "" {
		return nil, response.NewAPIError(http.StatusBadRequest, "Expert ID is required", "Expert ID parameter is missing")
	}

	pricingList, err := Expert().GetAllPriceByExpertID(ctx, expertID)
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "failed to get pricing config list", err.Error())
	}
	return pricingList, nil
}
func (ec *ExpertController) DeletePrice(ctx *gin.Context) (res interface{}, err error) {
	pricingID := ctx.Param("pricingId")
	if pricingID == "" {
		return nil, response.NewAPIError(http.StatusBadRequest, "Pricing ID is required", "Pricing ID parameter is missing")
	}

	err = Expert().DeletePrice(ctx, pricingID)
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "delete pricing config failed", err.Error())
	}
	return map[string]string{"message": "Pricing config deleted successfully"}, nil
}

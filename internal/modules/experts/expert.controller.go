package experts

import (
	dtoexperts "cbs_backend/internal/modules/experts/expertsdto"
	"cbs_backend/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
		return nil, response.NewAPIError(http.StatusBadRequest, "create expert is falied", err.Error())
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
		return nil, response.NewAPIError(http.StatusBadRequest, "update expert is falied", err.Error())
	}
	return resUpdateExpert, nil
}

func (ec *ExpertController) GetExpertProfileDetails(ctx *gin.Context) (res interface{}, err error) {
	idExpertStr := ctx.Param("id")
	expertID, err := uuid.Parse(idExpertStr)
	if err != nil {
		return nil, err
	}
	resExpertDetail, err := Expert().GetExpertProfileDetails(ctx, expertID)
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid response load detail", err.Error())
	}
	return resExpertDetail, nil
}

package expert

import (
	"cbs_backend/internal/middleware"
	PkgExpert "cbs_backend/internal/modules/experts"
	"cbs_backend/internal/modules/users"
	"cbs_backend/pkg/response"

	"github.com/gin-gonic/gin"
)

type ExpertRouter struct{}

func (er *ExpertRouter) InitExpertRouter(router *gin.RouterGroup) {
	expertCtrl := PkgExpert.NewExpertController()

	// Public routes - không cần authentication
	public := router.Group("/expert/v1")
	{
		public.GET("/getAllExpert", response.Wrap(expertCtrl.GetAllExpert))
		public.GET("/getDetail/:id", response.Wrap(expertCtrl.GetExpertProfileDetails))
		public.GET("/workHour/:expertId", response.Wrap(expertCtrl.GetAllWorkHourByExpertID))
		public.GET("/unavailableTime/:expertId", response.Wrap(expertCtrl.GetAllUnavailableTimeByExpertID))
		public.GET("/price/:expertId", response.Wrap(expertCtrl.GetAllPriceByExpertID))
		// 🆕 GET danh sách chuyên môn của chuyên gia
		public.GET("/specialization/:expertId", response.Wrap(expertCtrl.GetAllExpertSpecializationByExpertID))
	}

	// Private routes - cần authentication
	private := router.Group("/expert/v2")
	private.Use(middleware.AuthMiddleware(users.User()))
	{
		// Expert Profile Management
		private.POST("/createExpert", response.Wrap(expertCtrl.CreateExpertProfile))
		private.PUT("/update", response.Wrap(expertCtrl.UpdateExpertProfile))

		// Working Hours Management
		private.POST("/workHour", response.Wrap(expertCtrl.CreateWorkHour))
		private.PUT("/workHour", response.Wrap(expertCtrl.UpdateWorkHour))

		// Unavailable Time Management
		private.POST("/unavailableTime", response.Wrap(expertCtrl.CreateUnavailableTime))
		private.PUT("/unavailableTime", response.Wrap(expertCtrl.UpdateUnavailableTime))

		// 🆕 Specialization Management
		private.POST("/specialization", response.Wrap(expertCtrl.CreateExpertSpecialization))
		private.PUT("/specialization", response.Wrap(expertCtrl.UpdateExpertSpecialization))

		//Price Config
		private.POST("/price", response.Wrap(expertCtrl.CreatePrice))
		private.PUT("/price", response.Wrap(expertCtrl.UpdatePrice))
	}
}

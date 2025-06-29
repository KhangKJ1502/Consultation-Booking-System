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

	public := router.Group("/expert/v1")
	{
		public.GET("/getAllExpert", response.Wrap(expertCtrl.GetAllExpert))
		public.GET("/getDetail/:id", response.Wrap(expertCtrl.GetExpertProfileDetails))
	}

	private := router.Group("/expert/v2")
	private.Use(middleware.AuthMiddleware(users.User()))
	{
		private.POST("/createExpert", response.Wrap(expertCtrl.CreateExpertProfile))
		private.PUT("/update", response.Wrap(expertCtrl.UpdateExpertProfile))
	}
}

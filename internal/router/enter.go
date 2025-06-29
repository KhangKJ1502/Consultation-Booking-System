package router

import (
	"cbs_backend/internal/router/expert"
	"cbs_backend/internal/router/user"
)

type RouterGroup struct {
	User   user.RouterUserGroup
	Expert expert.RouterExpertGroup
}

var RouterGroupApp = new(RouterGroup)

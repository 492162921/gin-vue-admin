package inspection

import (
	api "github.com/flipped-aurora/gin-vue-admin/server/api/v1"
)

type RouterGroup struct {
	InspectionRouter
}

var (
	healthStubApi = api.ApiGroupApp.InspectionApiGroup.HealthStubApi
)

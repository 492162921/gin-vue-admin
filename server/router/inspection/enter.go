package inspection

import (
	api "github.com/flipped-aurora/gin-vue-admin/server/api/v1"
)

type RouterGroup struct {
	InspectionRouter
	ClusterRouter
	RuleRouter
	TaskRouter
}

var (
	healthStubApi = api.ApiGroupApp.InspectionApiGroup.HealthStubApi
	clusterApi    = api.ApiGroupApp.InspectionApiGroup.ClusterApi
	ruleApi       = api.ApiGroupApp.InspectionApiGroup.RuleApi
	taskApi       = api.ApiGroupApp.InspectionApiGroup.TaskApi
)

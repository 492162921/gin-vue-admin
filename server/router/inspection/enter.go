package inspection

import (
	api "github.com/flipped-aurora/gin-vue-admin/server/api/v1"
)

type RouterGroup struct {
	InspectionRouter
	ClusterRouter
	RuleRouter
	TaskRouter
	AlertRouter
	ReportRouter
	DashboardRouter
}

var (
	healthStubApi = api.ApiGroupApp.InspectionApiGroup.HealthStubApi
	clusterApi    = api.ApiGroupApp.InspectionApiGroup.ClusterApi
	ruleApi       = api.ApiGroupApp.InspectionApiGroup.RuleApi
	taskApi       = api.ApiGroupApp.InspectionApiGroup.TaskApi
	alertApi      = api.ApiGroupApp.InspectionApiGroup.AlertApi
	reportApi     = api.ApiGroupApp.InspectionApiGroup.ReportApi
	dashboardApi  = api.ApiGroupApp.InspectionApiGroup.DashboardApi
)

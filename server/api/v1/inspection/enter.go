package inspection

import service "github.com/flipped-aurora/gin-vue-admin/server/service"

type HealthStubApi struct{}

type ApiGroup struct {
	HealthStubApi
	ClusterApi
	RuleApi
	TaskApi
	AlertApi
	ReportApi
}

var (
	clusterService = service.ServiceGroupApp.InspectionServiceGroup.ClusterService
	ruleService    = service.ServiceGroupApp.InspectionServiceGroup.RuleService
	taskService    = service.ServiceGroupApp.InspectionServiceGroup.TaskService
	alertService   = service.ServiceGroupApp.InspectionServiceGroup.AlertService
	reportService  = service.ServiceGroupApp.InspectionServiceGroup.ReportService
)

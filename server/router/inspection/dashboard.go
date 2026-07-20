package inspection

import "github.com/gin-gonic/gin"

type DashboardRouter struct{}

func (r *DashboardRouter) InitDashboardRouter(router *gin.RouterGroup) {
	read := router.Group("inspection/dashboard")
	{
		read.GET("summary", dashboardApi.Summary)
		read.GET("resourceUsage", dashboardApi.ResourceUsage)
		read.GET("inspectionTrend", dashboardApi.InspectionTrend)
		read.GET("alertDistribution", dashboardApi.AlertDistribution)
	}
}

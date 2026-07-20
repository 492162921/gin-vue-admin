package inspection

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/gin-gonic/gin"
)

type DashboardApi struct{}

// Summary
// @Tags      InspDashboard
// @Summary   获取巡检总览统计
// @Security  ApiKeyAuth
// @Produce   application/json
// @Success   200  {object}  response.Response{data=inspection.DashboardSummary,msg=string}  "统计数据"
// @Router    /inspection/dashboard/summary [get]
func (a *DashboardApi) Summary(c *gin.Context) {
	summary, err := dashboardService.Summary(c.Request.Context())
	if err != nil {
		response.FailWithMessage("获取巡检总览失败", c)
		return
	}
	response.OkWithData(summary, c)
}

// ResourceUsage
// @Tags      InspDashboard
// @Summary   获取最近巡检的集群资源使用率
// @Security  ApiKeyAuth
// @Produce   application/json
// @Success   200  {object}  response.Response{data=[]inspection.DashboardResourceUsage,msg=string}  "资源使用率"
// @Router    /inspection/dashboard/resourceUsage [get]
func (a *DashboardApi) ResourceUsage(c *gin.Context) {
	usage, err := dashboardService.ResourceUsage(c.Request.Context())
	if err != nil {
		response.FailWithMessage("获取资源使用率失败", c)
		return
	}
	response.OkWithData(usage, c)
}

// InspectionTrend
// @Tags      InspDashboard
// @Summary   获取每日巡检异常趋势
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     days  query     int  false  "统计天数，默认 7，最大 90"
// @Success   200   {object}  response.Response{data=[]inspection.DashboardTrendItem,msg=string}  "异常趋势"
// @Router    /inspection/dashboard/inspectionTrend [get]
func (a *DashboardApi) InspectionTrend(c *gin.Context) {
	var query struct {
		Days int `form:"days"`
	}
	if err := c.ShouldBindQuery(&query); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	trend, err := dashboardService.InspectionTrend(c.Request.Context(), query.Days)
	if err != nil {
		response.FailWithMessage("获取巡检趋势失败", c)
		return
	}
	response.OkWithData(trend, c)
}

// AlertDistribution
// @Tags      InspDashboard
// @Summary   获取巡检告警分布
// @Security  ApiKeyAuth
// @Produce   application/json
// @Success   200  {object}  response.Response{data=inspection.DashboardAlertDistribution,msg=string}  "告警分布"
// @Router    /inspection/dashboard/alertDistribution [get]
func (a *DashboardApi) AlertDistribution(c *gin.Context) {
	distribution, err := dashboardService.AlertDistribution(c.Request.Context())
	if err != nil {
		response.FailWithMessage("获取告警分布失败", c)
		return
	}
	response.OkWithData(distribution, c)
}

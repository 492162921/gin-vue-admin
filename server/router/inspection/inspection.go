package inspection

import "github.com/gin-gonic/gin"

type InspectionRouter struct{}

func (r *InspectionRouter) InitInspectionRouter(Router *gin.RouterGroup) {
	g := Router.Group("inspection")
	{
		g.GET("ping", healthStubApi.Ping)
	}
	(&ClusterRouter{}).InitClusterRouter(Router)
	(&RuleRouter{}).InitRuleRouter(Router)
	(&TaskRouter{}).InitTaskRouter(Router)
	(&AlertRouter{}).InitAlertRouter(Router)
}

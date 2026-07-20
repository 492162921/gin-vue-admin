package inspection

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type RuleRouter struct{}

func (r *RuleRouter) InitRuleRouter(router *gin.RouterGroup) {
	record := router.Group("inspection/rule").Use(middleware.OperationRecord())
	read := router.Group("inspection/rule")
	{
		record.POST("", ruleApi.Create)
		record.PUT("", ruleApi.Update)
		record.DELETE("", ruleApi.Delete)
		record.POST("enabled", ruleApi.SetEnabled)
	}
	{
		read.GET("list", ruleApi.List)
	}
}

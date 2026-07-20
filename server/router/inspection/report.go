package inspection

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type ReportRouter struct{}

func (r *ReportRouter) InitReportRouter(router *gin.RouterGroup) {
	record := router.Group("inspection/report").Use(middleware.OperationRecord())
	read := router.Group("inspection/report")
	{
		record.POST("generate", reportApi.Generate)
		record.POST("push", reportApi.Push)
		record.DELETE("", reportApi.Delete)
	}
	{
		read.GET("", reportApi.Get)
		read.GET("list", reportApi.List)
	}
}

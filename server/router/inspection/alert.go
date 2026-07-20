package inspection

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type AlertRouter struct{}

func (r *AlertRouter) InitAlertRouter(router *gin.RouterGroup) {
	record := router.Group("inspection/alert").Use(middleware.OperationRecord())
	read := router.Group("inspection/alert")
	{
		record.PATCH("status", alertApi.UpdateStatus)
		record.DELETE("", alertApi.Delete)
		record.DELETE("batch-delete", alertApi.BatchDelete)
	}
	{
		read.GET("list", alertApi.List)
	}
}

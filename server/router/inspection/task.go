package inspection

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type TaskRouter struct{}

func (r *TaskRouter) InitTaskRouter(router *gin.RouterGroup) {
	record := router.Group("inspection/task").Use(middleware.OperationRecord())
	read := router.Group("inspection/task")
	{
		record.POST("", taskApi.Create)
		record.PUT("", taskApi.Update)
		record.DELETE("", taskApi.Delete)
		record.POST("run", taskApi.Run)
	}
	{
		read.GET("", taskApi.Get)
		read.GET("list", taskApi.List)
		read.GET("inspections", taskApi.Inspections)
	}
}

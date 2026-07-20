package inspection

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type ClusterRouter struct{}

func (r *ClusterRouter) InitClusterRouter(router *gin.RouterGroup) {
	record := router.Group("inspection/cluster").Use(middleware.OperationRecord())
	read := router.Group("inspection/cluster")
	{
		record.POST("", clusterApi.Create)
		record.PUT("", clusterApi.Update)
		record.DELETE("", clusterApi.Delete)
		record.POST("refresh", clusterApi.Refresh)
	}
	{
		read.GET("", clusterApi.Get)
		read.GET("list", clusterApi.List)
		read.GET("nodes", clusterApi.Nodes)
		read.GET("namespaces", clusterApi.Namespaces)
	}
}

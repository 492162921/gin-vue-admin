package inspection

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/gin-gonic/gin"
)

// Ping 提供巡检模块路由可达性检查，不验证任意 Kubernetes 集群状态。
// @Tags      InspHealth
// @Summary   巡检模块健康检查
// @Security  ApiKeyAuth
// @Produce   application/json
// @Success   200   {object}  response.Response{msg=string}  "pong"
// @Router    /inspection/ping [get]
func (h *HealthStubApi) Ping(c *gin.Context) {
	response.OkWithMessage("pong", c)
}

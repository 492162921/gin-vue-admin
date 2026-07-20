package inspection

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/gin-gonic/gin"
)

// Ping 临时健康占位，后续任务替换为真实 API
// @Tags      InspHealth
// @Summary   巡检模块健康检查
// @Security  ApiKeyAuth
// @Produce   application/json
// @Success   200   {object}  response.Response{msg=string}  "pong"
// @Router    /inspection/ping [get]
func (h *HealthStubApi) Ping(c *gin.Context) {
	response.OkWithMessage("pong", c)
}

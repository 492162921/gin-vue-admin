package inspection

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/gin-gonic/gin"
)

// Ping 临时健康占位，后续任务替换为真实 API
func (h *HealthStubApi) Ping(c *gin.Context) {
	response.OkWithMessage("pong", c)
}

package inspection

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	inspRequest "github.com/flipped-aurora/gin-vue-admin/server/model/inspection/request"
	"github.com/gin-gonic/gin"
)

type RuleApi struct{}

type setEnabledRequest struct {
	ID      uint `json:"id" binding:"required"`
	Enabled bool `json:"enabled"`
}

func (a *RuleApi) Create(c *gin.Context) {
	var input inspRequest.CreateRule
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	rule, err := ruleService.Create(c.Request.Context(), input)
	if err != nil {
		response.FailWithMessage("创建规则失败", c)
		return
	}
	response.OkWithData(rule, c)
}

func (a *RuleApi) Update(c *gin.Context) {
	var input inspRequest.UpdateRule
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	var id inspectionIDRequest
	if err := c.ShouldBindQuery(&id); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	rule, err := ruleService.Update(c.Request.Context(), id.ID, input)
	if err != nil {
		response.FailWithMessage("更新规则失败", c)
		return
	}
	response.OkWithData(rule, c)
}

func (a *RuleApi) Delete(c *gin.Context) {
	var input inspectionIDRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := ruleService.Delete(c.Request.Context(), input.ID); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

func (a *RuleApi) List(c *gin.Context) {
	var search inspRequest.RuleSearch
	if err := c.ShouldBindQuery(&search); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := ruleService.List(c.Request.Context(), search)
	if err != nil {
		response.FailWithMessage("获取规则列表失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: search.Page, PageSize: search.PageSize}, "获取成功", c)
}

func (a *RuleApi) SetEnabled(c *gin.Context) {
	var input setEnabledRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := ruleService.SetEnabled(c.Request.Context(), input.ID, input.Enabled); err != nil {
		response.FailWithMessage("更新规则状态失败", c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

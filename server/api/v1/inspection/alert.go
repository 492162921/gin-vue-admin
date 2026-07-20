package inspection

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	inspRequest "github.com/flipped-aurora/gin-vue-admin/server/model/inspection/request"
	"github.com/gin-gonic/gin"
)

type alertIDsRequest struct {
	IDs []uint `json:"ids" binding:"required"`
}

type AlertApi struct{}

// List
// @Tags      InspAlert
// @Summary   分页获取巡检告警
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     data  query     inspRequest.AlertSearch                                true  "分页与筛选条件"
// @Success   200   {object}  response.Response{data=response.PageResult{list=[]inspection.InspAlert},msg=string}  "告警列表"
// @Router    /inspection/alert/list [get]
func (a *AlertApi) List(c *gin.Context) {
	var search inspRequest.AlertSearch
	if err := c.ShouldBindQuery(&search); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := alertService.List(c.Request.Context(), search)
	if err != nil {
		response.FailWithMessage("获取巡检告警失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: search.Page, PageSize: search.PageSize}, "获取成功", c)
}

// UpdateStatus
// @Tags      InspAlert
// @Summary   更新巡检告警状态
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     id    query     uint                                                   true  "告警ID"
// @Param     data  body      inspRequest.PatchAlertStatus                           true  "告警状态"
// @Success   200   {object}  response.Response{msg=string}                         "更新状态"
// @Router    /inspection/alert/status [patch]
func (a *AlertApi) UpdateStatus(c *gin.Context) {
	var input inspRequest.PatchAlertStatus
	var id inspectionIDRequest
	if err := c.ShouldBindQuery(&id); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := alertService.UpdateStatus(c.Request.Context(), id.ID, input.Status); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// Delete
// @Tags      InspAlert
// @Summary   删除巡检告警
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      inspectionIDRequest             true  "告警ID"
// @Success   200   {object}  response.Response{msg=string}   "删除告警"
// @Router    /inspection/alert [delete]
func (a *AlertApi) Delete(c *gin.Context) {
	var input inspectionIDRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := alertService.Delete(c.Request.Context(), input.ID); err != nil {
		response.FailWithMessage("删除巡检告警失败", c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// BatchDelete
// @Tags      InspAlert
// @Summary   批量删除巡检告警
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      alertIDsRequest                  true  "告警ID列表"
// @Success   200   {object}  response.Response{msg=string}   "批量删除告警"
// @Router    /inspection/alert/batch-delete [delete]
func (a *AlertApi) BatchDelete(c *gin.Context) {
	var input alertIDsRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := alertService.DeleteByIDs(c.Request.Context(), input.IDs); err != nil {
		response.FailWithMessage("批量删除巡检告警失败", c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

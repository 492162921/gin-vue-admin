package inspection

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	inspRequest "github.com/flipped-aurora/gin-vue-admin/server/model/inspection/request"
	"github.com/gin-gonic/gin"
)

type ReportApi struct{}

// List
// @Tags      InspReport
// @Summary   分页获取巡检报告列表
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     data  query     inspRequest.ReportSearch                                  true  "分页与筛选条件"
// @Success   200   {object}  response.Response{data=response.PageResult{list=[]inspection.InspReport},msg=string}  "报告列表"
// @Router    /inspection/report/list [get]
func (a *ReportApi) List(c *gin.Context) {
	var search inspRequest.ReportSearch
	if err := c.ShouldBindQuery(&search); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := reportService.List(c.Request.Context(), search)
	if err != nil {
		response.FailWithMessage("获取巡检报告列表失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: search.Page, PageSize: search.PageSize}, "获取成功", c)
}

// Get
// @Tags      InspReport
// @Summary   获取巡检报告详情
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     id    query     uint                                                     true  "报告ID"
// @Success   200   {object}  response.Response{data=inspection.InspReport,msg=string}  "报告详情"
// @Router    /inspection/report [get]
func (a *ReportApi) Get(c *gin.Context) {
	var input inspectionIDRequest
	if err := c.ShouldBindQuery(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	report, err := reportService.Get(c.Request.Context(), input.ID)
	if err != nil {
		response.FailWithMessage("获取巡检报告失败", c)
		return
	}
	response.OkWithData(report, c)
}

// Generate
// @Tags      InspReport
// @Summary   为巡检结果生成报告
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      inspectionIDRequest                                      true  "巡检ID"
// @Success   200   {object}  response.Response{data=inspection.InspReport,msg=string}  "生成的报告"
// @Router    /inspection/report/generate [post]
func (a *ReportApi) Generate(c *gin.Context) {
	var input inspectionIDRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	report, err := reportService.Generate(c.Request.Context(), input.ID)
	if err != nil {
		response.FailWithMessage("生成巡检报告失败", c)
		return
	}
	response.OkWithData(report, c)
}

// Push
// @Tags      InspReport
// @Summary   推送巡检报告摘要到 Webhook
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      inspectionIDRequest              true  "报告ID"
// @Success   200   {object}  response.Response{msg=string}   "推送结果"
// @Router    /inspection/report/push [post]
func (a *ReportApi) Push(c *gin.Context) {
	var input inspectionIDRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := reportService.Push(c.Request.Context(), input.ID); err != nil {
		response.FailWithMessage("推送巡检报告失败", c)
		return
	}
	response.OkWithMessage("推送成功", c)
}

// Delete
// @Tags      InspReport
// @Summary   删除巡检报告
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      inspectionIDRequest             true  "报告ID"
// @Success   200   {object}  response.Response{msg=string}  "删除结果"
// @Router    /inspection/report [delete]
func (a *ReportApi) Delete(c *gin.Context) {
	var input inspectionIDRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := reportService.Delete(c.Request.Context(), input.ID); err != nil {
		response.FailWithMessage("删除巡检报告失败", c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

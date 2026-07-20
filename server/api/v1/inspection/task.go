package inspection

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	inspRequest "github.com/flipped-aurora/gin-vue-admin/server/model/inspection/request"
	"github.com/gin-gonic/gin"
)

type TaskApi struct{}

// Create
// @Tags      InspTask
// @Summary   创建巡检任务
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      inspRequest.CreateTask                                  true  "任务信息"
// @Success   200   {object}  response.Response{data=inspection.InspTask,msg=string}  "创建任务"
// @Router    /inspection/task [post]
func (a *TaskApi) Create(c *gin.Context) {
	var input inspRequest.CreateTask
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	task, err := taskService.Create(c.Request.Context(), input)
	if err != nil {
		response.FailWithMessage("创建巡检任务失败", c)
		return
	}
	response.OkWithData(task, c)
}

// Update
// @Tags      InspTask
// @Summary   更新巡检任务
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     id    query     uint                                                   true  "任务ID"
// @Param     data  body      inspRequest.UpdateTask                                 true  "任务信息"
// @Success   200   {object}  response.Response{data=inspection.InspTask,msg=string} "更新任务"
// @Router    /inspection/task [put]
func (a *TaskApi) Update(c *gin.Context) {
	var input inspRequest.UpdateTask
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	var id inspectionIDRequest
	if err := c.ShouldBindQuery(&id); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	task, err := taskService.Update(c.Request.Context(), id.ID, input)
	if err != nil {
		response.FailWithMessage("更新巡检任务失败", c)
		return
	}
	response.OkWithData(task, c)
}

// Delete
// @Tags      InspTask
// @Summary   删除巡检任务
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      inspectionIDRequest             true  "任务ID"
// @Success   200   {object}  response.Response{msg=string}   "删除任务"
// @Router    /inspection/task [delete]
func (a *TaskApi) Delete(c *gin.Context) {
	var input inspectionIDRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := taskService.Delete(c.Request.Context(), input.ID); err != nil {
		response.FailWithMessage("删除巡检任务失败", c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// Get
// @Tags      InspTask
// @Summary   获取巡检任务详情
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     id    query     uint                                                   true  "任务ID"
// @Success   200   {object}  response.Response{data=inspection.InspTask,msg=string} "任务详情"
// @Router    /inspection/task [get]
func (a *TaskApi) Get(c *gin.Context) {
	var input inspectionIDRequest
	if err := c.ShouldBindQuery(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	task, err := taskService.Get(c.Request.Context(), input.ID)
	if err != nil {
		response.FailWithMessage("获取巡检任务失败", c)
		return
	}
	response.OkWithData(task, c)
}

// List
// @Tags      InspTask
// @Summary   分页获取巡检任务列表
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     data  query     inspRequest.TaskSearch                              true  "分页与筛选条件"
// @Success   200   {object}  response.Response{data=response.PageResult,msg=string} "任务列表"
// @Router    /inspection/task/list [get]
func (a *TaskApi) List(c *gin.Context) {
	var search inspRequest.TaskSearch
	if err := c.ShouldBindQuery(&search); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := taskService.List(c.Request.Context(), search)
	if err != nil {
		response.FailWithMessage("获取巡检任务列表失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: search.Page, PageSize: search.PageSize}, "获取成功", c)
}

// Run
// @Tags      InspTask
// @Summary   立即执行巡检任务
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      inspectionIDRequest                                         true  "任务ID"
// @Success   200   {object}  response.Response{data=inspection.InspInspection,msg=string} "巡检结果"
// @Router    /inspection/task/run [post]
func (a *TaskApi) Run(c *gin.Context) {
	var input inspectionIDRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	inspection, err := taskService.RunNow(c.Request.Context(), input.ID)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(inspection, c)
}

// Inspections
// @Tags      InspTask
// @Summary   分页获取任务巡检历史
// @Security  ApiKeyAuth
// @Produce   application/json
// @Param     data  query     inspRequest.InspectionSearch                           true  "任务ID与分页条件"
// @Success   200   {object}  response.Response{data=response.PageResult,msg=string} "巡检历史"
// @Router    /inspection/task/inspections [get]
func (a *TaskApi) Inspections(c *gin.Context) {
	var search inspRequest.InspectionSearch
	if err := c.ShouldBindQuery(&search); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if search.TaskID == 0 {
		response.FailWithMessage("任务ID不能为空", c)
		return
	}
	list, total, err := taskService.ListInspections(c.Request.Context(), search.TaskID, search.PageInfo)
	if err != nil {
		response.FailWithMessage("获取巡检历史失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: search.Page, PageSize: search.PageSize}, "获取成功", c)
}

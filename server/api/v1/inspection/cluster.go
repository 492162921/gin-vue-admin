package inspection

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	inspRequest "github.com/flipped-aurora/gin-vue-admin/server/model/inspection/request"
	"github.com/gin-gonic/gin"
)

type ClusterApi struct{}

type inspectionIDRequest struct {
	ID uint `json:"id" form:"id" binding:"required"`
}

// Create
// @Tags      InspCluster
// @Summary   创建巡检集群
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      inspRequest.CreateCluster                                    true  "集群名称与 kubeconfig"
// @Success   200   {object}  response.Response{data=inspection.InspCluster,msg=string}      "创建集群"
// @Router    /inspection/cluster [post]
func (a *ClusterApi) Create(c *gin.Context) {
	var input inspRequest.CreateCluster
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	cluster, err := clusterService.Create(c.Request.Context(), input.Name, []byte(input.Kubeconfig))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(cluster, c)
}

// Update
// @Tags      InspCluster
// @Summary   更新巡检集群
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     id    query     uint                                                         true  "集群ID"
// @Param     data  body      inspRequest.UpdateCluster                                    true  "集群信息"
// @Success   200   {object}  response.Response{data=inspection.InspCluster,msg=string}     "更新集群"
// @Router    /inspection/cluster [put]
func (a *ClusterApi) Update(c *gin.Context) {
	var input inspRequest.UpdateCluster
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	var id inspectionIDRequest
	if err := c.ShouldBindQuery(&id); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	cluster, err := clusterService.Update(c.Request.Context(), id.ID, input)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(cluster, c)
}

// Delete
// @Tags      InspCluster
// @Summary   删除巡检集群
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      inspectionIDRequest                      true  "集群ID"
// @Success   200   {object}  response.Response{msg=string}            "删除集群"
// @Router    /inspection/cluster [delete]
func (a *ClusterApi) Delete(c *gin.Context) {
	var input inspectionIDRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := clusterService.Delete(c.Request.Context(), input.ID); err != nil {
		response.FailWithMessage("删除集群失败", c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// Get
// @Tags      InspCluster
// @Summary   获取巡检集群详情
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     id    query     uint                                                         true  "集群ID"
// @Success   200   {object}  response.Response{data=inspection.InspCluster,msg=string}     "获取集群详情"
// @Router    /inspection/cluster [get]
func (a *ClusterApi) Get(c *gin.Context) {
	var input inspectionIDRequest
	if err := c.ShouldBindQuery(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	cluster, err := clusterService.Get(c.Request.Context(), input.ID)
	if err != nil {
		response.FailWithMessage("获取集群失败", c)
		return
	}
	response.OkWithData(cluster, c)
}

// List
// @Tags      InspCluster
// @Summary   分页获取巡检集群列表
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  query     inspRequest.ClusterSearch                                    true  "分页与筛选条件"
// @Success   200   {object}  response.Response{data=response.PageResult,msg=string}       "分页获取集群列表"
// @Router    /inspection/cluster/list [get]
func (a *ClusterApi) List(c *gin.Context) {
	var search inspRequest.ClusterSearch
	if err := c.ShouldBindQuery(&search); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := clusterService.List(c.Request.Context(), search)
	if err != nil {
		response.FailWithMessage("获取集群列表失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: search.Page, PageSize: search.PageSize}, "获取成功", c)
}

// Refresh
// @Tags      InspCluster
// @Summary   刷新巡检集群状态
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      inspectionIDRequest                                          true  "集群ID"
// @Success   200   {object}  response.Response{data=inspection.InspCluster,msg=string}     "刷新集群状态"
// @Router    /inspection/cluster/refresh [post]
func (a *ClusterApi) Refresh(c *gin.Context) {
	var input inspectionIDRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	cluster, err := clusterService.Refresh(c.Request.Context(), input.ID)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(cluster, c)
}

// Nodes
// @Tags      InspCluster
// @Summary   获取集群节点列表
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     id    query     uint                                                         true  "集群ID"
// @Success   200   {object}  response.Response{data=[]k8s.NodeInfo,msg=string}            "获取节点列表"
// @Router    /inspection/cluster/nodes [get]
func (a *ClusterApi) Nodes(c *gin.Context) {
	var input inspectionIDRequest
	if err := c.ShouldBindQuery(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	nodes, err := clusterService.ListNodes(c.Request.Context(), input.ID)
	if err != nil {
		response.FailWithMessage("获取节点失败", c)
		return
	}
	response.OkWithData(nodes, c)
}

// Namespaces
// @Tags      InspCluster
// @Summary   获取集群命名空间列表
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     id    query     uint                                                         true  "集群ID"
// @Success   200   {object}  response.Response{data=[]k8s.NamespaceInfo,msg=string}       "获取命名空间列表"
// @Router    /inspection/cluster/namespaces [get]
func (a *ClusterApi) Namespaces(c *gin.Context) {
	var input inspectionIDRequest
	if err := c.ShouldBindQuery(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	namespaces, err := clusterService.ListNamespaces(c.Request.Context(), input.ID)
	if err != nil {
		response.FailWithMessage("获取命名空间失败", c)
		return
	}
	response.OkWithData(namespaces, c)
}

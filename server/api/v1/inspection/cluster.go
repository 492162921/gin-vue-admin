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

func (a *ClusterApi) Create(c *gin.Context) {
	var input inspRequest.CreateCluster
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	cluster, err := clusterService.Create(c.Request.Context(), input.Name, []byte(input.Kubeconfig))
	if err != nil {
		response.FailWithMessage("创建集群失败", c)
		return
	}
	response.OkWithData(cluster, c)
}

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
		response.FailWithMessage("更新集群失败", c)
		return
	}
	response.OkWithData(cluster, c)
}

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

func (a *ClusterApi) Refresh(c *gin.Context) {
	var input inspectionIDRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	cluster, err := clusterService.Refresh(c.Request.Context(), input.ID)
	if err != nil {
		response.FailWithMessage("刷新集群失败", c)
		return
	}
	response.OkWithData(cluster, c)
}

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

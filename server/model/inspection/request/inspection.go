package request

import "github.com/flipped-aurora/gin-vue-admin/server/model/common/request"

type ClusterSearch struct {
	request.PageInfo
	Name   string `json:"name" form:"name"`
	Status string `json:"status" form:"status"`
}

type CreateCluster struct {
	Name       string `json:"name" binding:"required"`
	Kubeconfig string `json:"kubeconfig" binding:"required"`
}

type UpdateCluster struct {
	Name       string `json:"name" binding:"required"`
	Kubeconfig string `json:"kubeconfig"`
}

type RuleSearch struct {
	request.PageInfo
	Name     string `json:"name" form:"name"`
	RuleType string `json:"ruleType" form:"ruleType"`
	Enabled  *bool  `json:"enabled" form:"enabled"`
}

type CreateRule struct {
	Name      string  `json:"name" binding:"required"`
	RuleType  string  `json:"ruleType" binding:"required"`
	Threshold float64 `json:"threshold"`
	Scope     string  `json:"scope"`
	Namespace string  `json:"namespace"`
}

type UpdateRule struct {
	Name      string  `json:"name" binding:"required"`
	RuleType  string  `json:"ruleType" binding:"required"`
	Threshold float64 `json:"threshold"`
	Scope     string  `json:"scope"`
	Namespace string  `json:"namespace"`
	Enabled   *bool   `json:"enabled"`
}

type TaskSearch struct {
	request.PageInfo
	Name      string `json:"name" form:"name"`
	ClusterID uint   `json:"clusterId" form:"clusterId"`
	Status    string `json:"status" form:"status"`
}

type CreateTask struct {
	Name      string `json:"name" binding:"required"`
	ClusterID uint   `json:"clusterId" binding:"required"`
	CronExpr  string `json:"cronExpr"`
	RuleIDs   []uint `json:"ruleIds" binding:"required"`
}

type UpdateTask struct {
	Name      string `json:"name" binding:"required"`
	ClusterID uint   `json:"clusterId" binding:"required"`
	CronExpr  string `json:"cronExpr"`
	RuleIDs   []uint `json:"ruleIds" binding:"required"`
	Status    string `json:"status"`
}

type AlertSearch struct {
	request.PageInfo
	Status   string `json:"status" form:"status"`
	Level    string `json:"level" form:"level"`
	RuleType string `json:"ruleType" form:"ruleType"`
}

type PatchAlertStatus struct {
	Status string `json:"status" binding:"required"`
}

type ReportSearch struct {
	request.PageInfo
	Title        string `json:"title" form:"title"`
	InspectionID uint   `json:"inspectionId" form:"inspectionId"`
}

type InspectionSearch struct {
	request.PageInfo
	TaskID uint   `json:"taskId" form:"taskId"`
	Status string `json:"status" form:"status"`
}

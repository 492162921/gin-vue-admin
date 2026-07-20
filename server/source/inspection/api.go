package inspection

import (
	"context"

	sysModel "github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/flipped-aurora/gin-vue-admin/server/service/system"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

const initOrderAPI = system.InitOrderExternal + 3

type initAPI struct{}

func init() {
	system.RegisterInit(initOrderAPI, &initAPI{})
}

func (i *initAPI) InitializerName() string { return "inspection_apis" }

func (i *initAPI) MigrateTable(ctx context.Context) (context.Context, error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return ctx, system.ErrMissingDBContext
	}
	return ctx, db.AutoMigrate(&sysModel.SysApi{})
}

func (i *initAPI) TableCreated(ctx context.Context) bool {
	db, ok := ctx.Value("db").(*gorm.DB)
	return ok && db.Migrator().HasTable(&sysModel.SysApi{})
}

func (i *initAPI) InitializeData(ctx context.Context) (context.Context, error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return ctx, system.ErrMissingDBContext
	}
	apis := inspectionAPIs()
	for _, api := range apis {
		if err := db.Where("path = ? AND method = ?", api.Path, api.Method).FirstOrCreate(&api).Error; err != nil {
			return ctx, errors.Wrap(err, "初始化巡检 API 失败")
		}
	}
	return ctx, nil
}

func (i *initAPI) DataInserted(ctx context.Context) bool {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return false
	}
	var count int64
	return db.Model(&sysModel.SysApi{}).Where("api_group = ?", "K8s巡检").Count(&count).Error == nil && count == int64(len(inspectionAPIs()))
}

func inspectionAPIs() []sysModel.SysApi {
	return []sysModel.SysApi{
		{ApiGroup: "K8s巡检", Method: "GET", Path: "/inspection/ping", Description: "获取巡检服务状态"},
		{ApiGroup: "K8s巡检", Method: "GET", Path: "/inspection/dashboard/summary", Description: "获取巡检总览统计"},
		{ApiGroup: "K8s巡检", Method: "GET", Path: "/inspection/dashboard/resourceUsage", Description: "获取集群资源使用率"},
		{ApiGroup: "K8s巡检", Method: "GET", Path: "/inspection/dashboard/inspectionTrend", Description: "获取巡检异常趋势"},
		{ApiGroup: "K8s巡检", Method: "GET", Path: "/inspection/dashboard/alertDistribution", Description: "获取巡检告警分布"},
		{ApiGroup: "K8s巡检", Method: "POST", Path: "/inspection/cluster", Description: "创建巡检集群"},
		{ApiGroup: "K8s巡检", Method: "PUT", Path: "/inspection/cluster", Description: "更新巡检集群"},
		{ApiGroup: "K8s巡检", Method: "DELETE", Path: "/inspection/cluster", Description: "删除巡检集群"},
		{ApiGroup: "K8s巡检", Method: "GET", Path: "/inspection/cluster", Description: "获取巡检集群"},
		{ApiGroup: "K8s巡检", Method: "GET", Path: "/inspection/cluster/list", Description: "获取巡检集群列表"},
		{ApiGroup: "K8s巡检", Method: "GET", Path: "/inspection/cluster/nodes", Description: "获取集群节点"},
		{ApiGroup: "K8s巡检", Method: "GET", Path: "/inspection/cluster/namespaces", Description: "获取集群命名空间"},
		{ApiGroup: "K8s巡检", Method: "POST", Path: "/inspection/cluster/refresh", Description: "刷新集群状态"},
		{ApiGroup: "K8s巡检", Method: "POST", Path: "/inspection/rule", Description: "创建巡检规则"},
		{ApiGroup: "K8s巡检", Method: "PUT", Path: "/inspection/rule", Description: "更新巡检规则"},
		{ApiGroup: "K8s巡检", Method: "DELETE", Path: "/inspection/rule", Description: "删除巡检规则"},
		{ApiGroup: "K8s巡检", Method: "GET", Path: "/inspection/rule/list", Description: "获取巡检规则列表"},
		{ApiGroup: "K8s巡检", Method: "POST", Path: "/inspection/rule/enabled", Description: "启停巡检规则"},
		{ApiGroup: "K8s巡检", Method: "POST", Path: "/inspection/task", Description: "创建巡检任务"},
		{ApiGroup: "K8s巡检", Method: "PUT", Path: "/inspection/task", Description: "更新巡检任务"},
		{ApiGroup: "K8s巡检", Method: "DELETE", Path: "/inspection/task", Description: "删除巡检任务"},
		{ApiGroup: "K8s巡检", Method: "GET", Path: "/inspection/task", Description: "获取巡检任务"},
		{ApiGroup: "K8s巡检", Method: "GET", Path: "/inspection/task/list", Description: "获取巡检任务列表"},
		{ApiGroup: "K8s巡检", Method: "POST", Path: "/inspection/task/run", Description: "立即执行巡检任务"},
		{ApiGroup: "K8s巡检", Method: "GET", Path: "/inspection/task/inspections", Description: "获取巡检任务历史"},
		{ApiGroup: "K8s巡检", Method: "GET", Path: "/inspection/alert/list", Description: "获取巡检告警列表"},
		{ApiGroup: "K8s巡检", Method: "PATCH", Path: "/inspection/alert/status", Description: "更新巡检告警状态"},
		{ApiGroup: "K8s巡检", Method: "DELETE", Path: "/inspection/alert", Description: "删除巡检告警"},
		{ApiGroup: "K8s巡检", Method: "DELETE", Path: "/inspection/alert/batch-delete", Description: "批量删除巡检告警"},
		{ApiGroup: "K8s巡检", Method: "GET", Path: "/inspection/report/list", Description: "获取巡检报告列表"},
		{ApiGroup: "K8s巡检", Method: "GET", Path: "/inspection/report", Description: "获取巡检报告详情"},
		{ApiGroup: "K8s巡检", Method: "POST", Path: "/inspection/report/generate", Description: "生成巡检报告"},
		{ApiGroup: "K8s巡检", Method: "POST", Path: "/inspection/report/push", Description: "推送巡检报告"},
		{ApiGroup: "K8s巡检", Method: "DELETE", Path: "/inspection/report", Description: "删除巡检报告"},
	}
}

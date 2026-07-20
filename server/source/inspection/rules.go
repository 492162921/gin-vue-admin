package inspection

import (
	"context"

	inspModel "github.com/flipped-aurora/gin-vue-admin/server/model/inspection"
	"github.com/flipped-aurora/gin-vue-admin/server/service/system"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

const initOrderRules = system.InitOrderExternal + 1

type initRules struct{}

func init() {
	system.RegisterInit(initOrderRules, &initRules{})
}

func (i *initRules) InitializerName() string { return "inspection_rules" }

func (i *initRules) MigrateTable(ctx context.Context) (context.Context, error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return ctx, system.ErrMissingDBContext
	}
	return ctx, db.AutoMigrate(&inspModel.InspRule{})
}

func (i *initRules) TableCreated(ctx context.Context) bool {
	db, ok := ctx.Value("db").(*gorm.DB)
	return ok && db.Migrator().HasTable(&inspModel.InspRule{})
}

func (i *initRules) InitializeData(ctx context.Context) (context.Context, error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return ctx, system.ErrMissingDBContext
	}
	rules := []inspModel.InspRule{
		{Name: "节点 CPU 使用率", RuleType: "node_cpu", Threshold: 80, Scope: "cluster", Enabled: true},
		{Name: "节点内存使用率", RuleType: "node_memory", Threshold: 80, Scope: "cluster", Enabled: true},
		{Name: "Pod 重启次数", RuleType: "pod_restart", Threshold: 1, Scope: "cluster", Enabled: true},
		{Name: "Pod 高重启次数", RuleType: "pod_restart_high", Threshold: 5, Scope: "cluster", Enabled: true},
		{Name: "Pod 未运行", RuleType: "pod_not_running", Scope: "cluster", Enabled: true},
		{Name: "节点未就绪", RuleType: "node_not_ready", Scope: "cluster", Enabled: true},
		{Name: "节点磁盘压力", RuleType: "node_disk_pressure", Scope: "cluster", Enabled: true},
		{Name: "节点内存压力", RuleType: "node_memory_pressure", Scope: "cluster", Enabled: true},
		{Name: "节点 PID 压力", RuleType: "node_pid_pressure", Scope: "cluster", Enabled: true},
		{Name: "节点网络不可用", RuleType: "node_network_unavailable", Scope: "cluster", Enabled: true},
		{Name: "节点不可调度", RuleType: "node_unschedulable", Scope: "cluster", Enabled: true},
		{Name: "Pod Pending", RuleType: "pod_pending", Scope: "cluster", Enabled: true},
		{Name: "Pod Failed", RuleType: "pod_failed", Scope: "cluster", Enabled: true},
		{Name: "Pod Unknown", RuleType: "pod_unknown", Scope: "cluster", Enabled: true},
		{Name: "Pod OOMKilled", RuleType: "pod_oomkilled", Scope: "cluster", Enabled: true},
		{Name: "镜像拉取失败", RuleType: "pod_image_pull_backoff", Scope: "cluster", Enabled: true},
		{Name: "Pod CrashLoop", RuleType: "pod_crash_loop", Scope: "cluster", Enabled: true},
		{Name: "Pod 未就绪", RuleType: "pod_not_ready", Scope: "cluster", Enabled: true},
		{Name: "Pod 已驱逐", RuleType: "pod_evicted", Scope: "cluster", Enabled: true},
		{Name: "容器未就绪", RuleType: "container_not_ready", Scope: "cluster", Enabled: true},
	}
	for _, rule := range rules {
		if err := db.Where("rule_type = ?", rule.RuleType).FirstOrCreate(&rule).Error; err != nil {
			return ctx, errors.Wrap(err, "初始化巡检规则失败")
		}
	}
	return ctx, nil
}

func (i *initRules) DataInserted(ctx context.Context) bool {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return false
	}
	var count int64
	return db.Model(&inspModel.InspRule{}).Where("rule_type IN ?", []string{
		"node_cpu", "node_memory", "pod_restart", "pod_restart_high", "pod_not_running",
		"node_not_ready", "node_disk_pressure", "node_memory_pressure", "node_pid_pressure",
		"node_network_unavailable", "node_unschedulable", "pod_pending", "pod_failed",
		"pod_unknown", "pod_oomkilled", "pod_image_pull_backoff", "pod_crash_loop",
		"pod_not_ready", "pod_evicted", "container_not_ready",
	}).Count(&count).Error == nil && count == 20
}

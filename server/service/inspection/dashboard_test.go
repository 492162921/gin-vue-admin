package inspection

import (
	"context"
	"testing"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/internal/testutil"
	inspModel "github.com/flipped-aurora/gin-vue-admin/server/model/inspection"
)

func TestDashboardServiceReturnsSeededMetrics(t *testing.T) {
	now := time.Now()
	testutil.NewMemoryDB(t,
		&inspModel.InspCluster{},
		&inspModel.InspTask{},
		&inspModel.InspInspection{},
		&inspModel.InspInspectionDetail{},
		&inspModel.InspAlert{},
	)
	seedDashboardData(t, now)

	service := DashboardService{}
	summary, err := service.Summary(context.Background())
	if err != nil {
		t.Fatalf("summary: %v", err)
	}
	if summary.ClusterCount != 2 || summary.NodeCount != 5 || summary.TodayInspectionCount != 1 || summary.OpenAlertCount != 1 {
		t.Fatalf("summary = %#v", summary)
	}

	usage, err := service.ResourceUsage(context.Background())
	if err != nil {
		t.Fatalf("resource usage: %v", err)
	}
	if len(usage) != 1 || usage[0].ClusterName != "生产集群" || usage[0].CPUUsage != 70 || usage[0].MemoryUsage != 60 {
		t.Fatalf("resource usage = %#v", usage)
	}

	trend, err := service.InspectionTrend(context.Background(), 3)
	if err != nil {
		t.Fatalf("inspection trend: %v", err)
	}
	if len(trend) != 3 || trend[2].AnomalyCount != 3 || trend[1].AnomalyCount != 2 || trend[0].AnomalyCount != 0 {
		t.Fatalf("inspection trend = %#v", trend)
	}

	distribution, err := service.AlertDistribution(context.Background())
	if err != nil {
		t.Fatalf("alert distribution: %v", err)
	}
	if len(distribution.ByLevel) != 2 || distribution.ByLevel[0].Name != "critical" || distribution.ByLevel[0].Count != 1 {
		t.Fatalf("alert level distribution = %#v", distribution.ByLevel)
	}
	if len(distribution.ByCluster) != 1 || distribution.ByCluster[0].Name != "生产集群" || distribution.ByCluster[0].Count != 2 {
		t.Fatalf("alert cluster distribution = %#v", distribution.ByCluster)
	}
}

func seedDashboardData(t *testing.T, now time.Time) {
	t.Helper()
	cluster := inspModel.InspCluster{Name: "生产集群", NodeCount: 3}
	emptyCluster := inspModel.InspCluster{Name: "测试集群", NodeCount: 2}
	for _, item := range []*inspModel.InspCluster{&cluster, &emptyCluster} {
		if err := global.GVA_DB.Create(item).Error; err != nil {
			t.Fatalf("create cluster: %v", err)
		}
	}
	task := inspModel.InspTask{Name: "生产巡检", ClusterID: cluster.ID}
	if err := global.GVA_DB.Create(&task).Error; err != nil {
		t.Fatalf("create task: %v", err)
	}
	inspection := inspModel.InspInspection{TaskID: task.ID, Status: "success", AnomalyCount: 3, StartedAt: now, FinishedAt: &now}
	previous := inspModel.InspInspection{TaskID: task.ID, Status: "success", AnomalyCount: 2, StartedAt: now.AddDate(0, 0, -1)}
	if err := global.GVA_DB.Create(&inspection).Error; err != nil {
		t.Fatalf("create inspection: %v", err)
	}
	if err := global.GVA_DB.Create(&previous).Error; err != nil {
		t.Fatalf("create previous inspection: %v", err)
	}
	details := []inspModel.InspInspectionDetail{
		{InspectionID: inspection.ID, RuleType: "node_cpu", MetricValue: 60},
		{InspectionID: inspection.ID, RuleType: "node_cpu", MetricValue: 80},
		{InspectionID: inspection.ID, RuleType: "node_memory", MetricValue: 50},
		{InspectionID: inspection.ID, RuleType: "node_memory", MetricValue: 70},
	}
	if err := global.GVA_DB.Create(&details).Error; err != nil {
		t.Fatalf("create details: %v", err)
	}
	alerts := []inspModel.InspAlert{
		{InspectionID: inspection.ID, Level: "critical", Status: "open", CollectedAt: now},
		{InspectionID: inspection.ID, Level: "warning", Status: "closed", CollectedAt: now},
	}
	if err := global.GVA_DB.Create(&alerts).Error; err != nil {
		t.Fatalf("create alerts: %v", err)
	}
}

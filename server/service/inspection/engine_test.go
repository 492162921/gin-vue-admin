package inspection

import (
	"context"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/internal/testutil"
	inspModel "github.com/flipped-aurora/gin-vue-admin/server/model/inspection"
	"github.com/flipped-aurora/gin-vue-admin/server/service/inspection/k8s"
)

func TestEngineRunWithStubMarksAnomalies(t *testing.T) {
	db := testutil.NewMemoryDB(t,
		&inspModel.InspCluster{},
		&inspModel.InspRule{},
		&inspModel.InspTask{},
		&inspModel.InspTaskRule{},
		&inspModel.InspInspection{},
		&inspModel.InspInspectionDetail{},
	)
	cluster := inspModel.InspCluster{Name: "stub 集群"}
	rules := []inspModel.InspRule{
		{Name: "节点未就绪", RuleType: "node_not_ready", Enabled: true},
		{Name: "Pod 非运行", RuleType: "pod_not_running", Enabled: true},
	}
	task := inspModel.InspTask{Name: "巡检任务", Cluster: &cluster, Rules: rules, Status: "active"}
	if err := db.Create(&task).Error; err != nil {
		t.Fatalf("创建巡检任务失败: %v", err)
	}

	engine := NewEngine(func(inspModel.InspCluster) k8s.Inspector {
		return k8s.NewInspector("", true)
	})
	inspection, err := engine.Run(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("执行巡检失败: %v", err)
	}
	if inspection.AnomalyCount == 0 {
		t.Fatal("Stub 巡检应记录异常")
	}
	var details []inspModel.InspInspectionDetail
	if err := db.Where("inspection_id = ?", inspection.ID).Find(&details).Error; err != nil {
		t.Fatalf("读取巡检明细失败: %v", err)
	}
	if len(details) == 0 {
		t.Fatal("巡检明细未落库")
	}
}

func TestEngineRespectsNamespaceScope(t *testing.T) {
	db := testutil.NewMemoryDB(t,
		&inspModel.InspCluster{},
		&inspModel.InspRule{},
		&inspModel.InspTask{},
		&inspModel.InspTaskRule{},
		&inspModel.InspInspection{},
		&inspModel.InspInspectionDetail{},
	)
	cluster := inspModel.InspCluster{Name: "测试集群"}
	rule := inspModel.InspRule{
		Name: "指定命名空间 Pod 非运行", RuleType: "pod_not_running",
		Scope: "namespace", Namespace: "ns-a", Enabled: true,
	}
	task := inspModel.InspTask{Name: "命名空间巡检", Cluster: &cluster, Rules: []inspModel.InspRule{rule}}
	if err := db.Create(&task).Error; err != nil {
		t.Fatalf("创建巡检任务失败: %v", err)
	}

	engine := NewEngine(func(inspModel.InspCluster) k8s.Inspector {
		return namespaceInspector{}
	})
	inspection, err := engine.Run(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("执行巡检失败: %v", err)
	}
	var details []inspModel.InspInspectionDetail
	if err := db.Where("inspection_id = ?", inspection.ID).Find(&details).Error; err != nil {
		t.Fatalf("读取巡检明细失败: %v", err)
	}
	if len(details) != 1 || details[0].ResourceName != "ns-a/bad-pod" {
		t.Fatalf("命名空间规则只能产生 ns-a 明细，实际为: %#v", details)
	}
}

type namespaceInspector struct{}

func (namespaceInspector) TestConnection(context.Context) error { return nil }
func (namespaceInspector) GetVersion(context.Context) (string, error) {
	return "v1.test", nil
}
func (namespaceInspector) ListNodes(context.Context) ([]k8s.NodeInfo, error) { return nil, nil }
func (namespaceInspector) ListPods(_ context.Context, namespace string) ([]k8s.PodInfo, error) {
	if namespace == "ns-a" {
		return []k8s.PodInfo{{Namespace: "ns-a", Name: "bad-pod", Phase: "Failed"}}, nil
	}
	return []k8s.PodInfo{
		{Namespace: "ns-a", Name: "bad-pod", Phase: "Failed"},
		{Namespace: "ns-b", Name: "other-bad-pod", Phase: "Failed"},
	}, nil
}
func (namespaceInspector) ListNamespaces(context.Context) ([]k8s.NamespaceInfo, error) {
	return nil, nil
}

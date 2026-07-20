package inspection

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	inspModel "github.com/flipped-aurora/gin-vue-admin/server/model/inspection"
	"github.com/flipped-aurora/gin-vue-admin/server/service/inspection/k8s"
)

type InspectorFactory func(inspModel.InspCluster) k8s.Inspector

type Engine struct {
	inspectorFactory InspectorFactory
	running          sync.Map
}

func NewEngine(factory InspectorFactory) *Engine {
	if factory == nil {
		factory = func(cluster inspModel.InspCluster) k8s.Inspector {
			return k8s.NewInspector(cluster.KubeconfigPath, global.GVA_CONFIG.Inspection.ForceStub)
		}
	}
	return &Engine{inspectorFactory: factory}
}

func (e *Engine) Run(ctx context.Context, taskID uint) (*inspModel.InspInspection, error) {
	if _, loaded := e.running.LoadOrStore(taskID, struct{}{}); loaded {
		return nil, fmt.Errorf("巡检任务 %d 正在执行", taskID)
	}
	defer e.running.Delete(taskID)

	var task inspModel.InspTask
	if err := global.GVA_DB.WithContext(ctx).Preload("Cluster").Preload("Rules").First(&task, taskID).Error; err != nil {
		return nil, err
	}
	if task.Cluster == nil {
		return nil, fmt.Errorf("巡检任务未关联集群")
	}

	inspection := &inspModel.InspInspection{TaskID: taskID, Status: "running", StartedAt: time.Now()}
	if err := global.GVA_DB.WithContext(ctx).Create(inspection).Error; err != nil {
		return nil, err
	}
	inspector := e.inspectorFactory(*task.Cluster)
	details, err := evaluateRules(ctx, inspector, task.Rules, inspection.ID)
	if err != nil {
		return nil, e.failInspection(ctx, inspection, err)
	}
	if len(details) > 0 {
		if err := global.GVA_DB.WithContext(ctx).Create(&details).Error; err != nil {
			return nil, e.failInspection(ctx, inspection, err)
		}
	}

	anomalyCount := 0
	for _, detail := range details {
		if detail.IsAnomaly {
			anomalyCount++
		}
	}
	now := time.Now()
	inspection.Status = "success"
	inspection.AnomalyCount = anomalyCount
	inspection.Summary = fmt.Sprintf("巡检完成，共 %d 项，异常 %d 项", len(details), anomalyCount)
	inspection.FinishedAt = &now
	if err := global.GVA_DB.WithContext(ctx).Save(inspection).Error; err != nil {
		return nil, err
	}
	if anomalyCount > 0 {
		alerts, samples := alertsFromDetails(inspection.ID, details)
		if err := defaultAlertService.Create(ctx, alerts); err != nil {
			return nil, err
		}
		_ = NotifyWebhook(global.GVA_CONFIG.Inspection.WebhookURL, AlertWebhookPayload{
			TaskID:       task.ID,
			TaskName:     task.Name,
			ClusterName:  task.Cluster.Name,
			InspectionID: inspection.ID,
			AnomalyCount: anomalyCount,
			Summary:      inspection.Summary,
			Samples:      samples,
		})
	}
	return inspection, nil
}

func (e *Engine) failInspection(ctx context.Context, inspection *inspModel.InspInspection, cause error) error {
	now := time.Now()
	inspection.Status = "failed"
	inspection.Summary = fmt.Sprintf("巡检失败：%v", cause)
	inspection.FinishedAt = &now
	if err := global.GVA_DB.WithContext(ctx).Save(inspection).Error; err != nil {
		return fmt.Errorf("%w; 更新巡检失败状态失败: %v", cause, err)
	}
	return cause
}

func evaluateRules(ctx context.Context, inspector k8s.Inspector, rules []inspModel.InspRule, inspectionID uint) ([]inspModel.InspInspectionDetail, error) {
	var details []inspModel.InspInspectionDetail
	var nodes []k8s.NodeInfo
	nodesLoaded := false
	podsByNamespace := map[string][]k8s.PodInfo{}

	loadNodes := func() ([]k8s.NodeInfo, error) {
		if !nodesLoaded {
			var err error
			nodes, err = inspector.ListNodes(ctx)
			if err != nil {
				return nil, fmt.Errorf("获取节点失败: %w", err)
			}
			nodesLoaded = true
		}
		return nodes, nil
	}
	loadPods := func(namespace string) ([]k8s.PodInfo, error) {
		if pods, ok := podsByNamespace[namespace]; ok {
			return pods, nil
		}
		pods, err := inspector.ListPods(ctx, namespace)
		if err != nil {
			return nil, fmt.Errorf("获取 Pod 失败: %w", err)
		}
		podsByNamespace[namespace] = pods
		return pods, nil
	}
	appendNodeDetails := func(rule inspModel.InspRule, message func(k8s.NodeInfo) string, metric func(k8s.NodeInfo) float64, anomalous func(k8s.NodeInfo) bool) error {
		nodes, err := loadNodes()
		if err != nil {
			return err
		}
		for _, node := range nodes {
			details = append(details, inspModel.InspInspectionDetail{
				InspectionID: inspectionID, RuleType: rule.RuleType, ResourceType: "node",
				ResourceName: node.Name, Message: message(node), MetricValue: metric(node), IsAnomaly: anomalous(node),
			})
		}
		return nil
	}
	appendPodDetails := func(rule inspModel.InspRule, message func(k8s.PodInfo) string, metric func(k8s.PodInfo) float64, anomalous func(k8s.PodInfo) bool) error {
		namespace := ""
		if rule.Scope == "namespace" {
			namespace = rule.Namespace
		}
		pods, err := loadPods(namespace)
		if err != nil {
			return err
		}
		for _, pod := range pods {
			details = append(details, inspModel.InspInspectionDetail{
				InspectionID: inspectionID, RuleType: rule.RuleType, ResourceType: "pod",
				ResourceName: pod.Namespace + "/" + pod.Name, Message: message(pod), MetricValue: metric(pod), IsAnomaly: anomalous(pod),
			})
		}
		return nil
	}

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}
		switch normalizeRuleType(rule.RuleType) {
		case "node_cpu":
			if err := appendNodeDetails(rule, func(n k8s.NodeInfo) string { return fmt.Sprintf("CPU %.1f%%", n.CPUUsage) }, func(n k8s.NodeInfo) float64 { return n.CPUUsage }, func(n k8s.NodeInfo) bool { return n.CPUUsage > rule.Threshold }); err != nil {
				return nil, err
			}
		case "node_memory":
			if err := appendNodeDetails(rule, func(n k8s.NodeInfo) string { return fmt.Sprintf("Memory %.1f%%", n.MemUsage) }, func(n k8s.NodeInfo) float64 { return n.MemUsage }, func(n k8s.NodeInfo) bool { return n.MemUsage > rule.Threshold }); err != nil {
				return nil, err
			}
		case "node_not_ready":
			if err := appendNodeDetails(rule, func(n k8s.NodeInfo) string { return fmt.Sprintf("Status=%s", n.Status) }, func(k8s.NodeInfo) float64 { return 0 }, func(n k8s.NodeInfo) bool { return n.Status != "Ready" }); err != nil {
				return nil, err
			}
		case "node_disk_pressure":
			if err := appendNodeDetails(rule, func(n k8s.NodeInfo) string { return fmt.Sprintf("DiskPressure=%t", n.DiskPressure) }, func(k8s.NodeInfo) float64 { return 0 }, func(n k8s.NodeInfo) bool { return n.DiskPressure }); err != nil {
				return nil, err
			}
		case "node_memory_pressure":
			if err := appendNodeDetails(rule, func(n k8s.NodeInfo) string { return fmt.Sprintf("MemoryPressure=%t", n.MemoryPressure) }, func(k8s.NodeInfo) float64 { return 0 }, func(n k8s.NodeInfo) bool { return n.MemoryPressure }); err != nil {
				return nil, err
			}
		case "node_pid_pressure":
			if err := appendNodeDetails(rule, func(n k8s.NodeInfo) string { return fmt.Sprintf("PIDPressure=%t", n.PIDPressure) }, func(k8s.NodeInfo) float64 { return 0 }, func(n k8s.NodeInfo) bool { return n.PIDPressure }); err != nil {
				return nil, err
			}
		case "node_network_unavailable":
			if err := appendNodeDetails(rule, func(n k8s.NodeInfo) string { return fmt.Sprintf("NetworkUnavailable=%t", n.NetworkUnavailable) }, func(k8s.NodeInfo) float64 { return 0 }, func(n k8s.NodeInfo) bool { return n.NetworkUnavailable }); err != nil {
				return nil, err
			}
		case "node_unschedulable":
			if err := appendNodeDetails(rule, func(n k8s.NodeInfo) string { return fmt.Sprintf("Unschedulable=%t", n.Unschedulable) }, func(k8s.NodeInfo) float64 { return 0 }, func(n k8s.NodeInfo) bool { return n.Unschedulable }); err != nil {
				return nil, err
			}
		case "pod_restart", "pod_restart_high":
			if err := appendPodDetails(rule, func(p k8s.PodInfo) string { return fmt.Sprintf("Restarts %d", p.RestartCount) }, func(p k8s.PodInfo) float64 { return float64(p.RestartCount) }, func(p k8s.PodInfo) bool { return float64(p.RestartCount) > rule.Threshold }); err != nil {
				return nil, err
			}
		case "pod_not_running":
			if err := appendPodDetails(rule, func(p k8s.PodInfo) string { return "Phase " + p.Phase }, func(k8s.PodInfo) float64 { return 0 }, func(p k8s.PodInfo) bool { return p.Phase != "Running" && p.Phase != "Succeeded" }); err != nil {
				return nil, err
			}
		case "pod_pending":
			if err := appendPodDetails(rule, func(k8s.PodInfo) string { return "Phase Pending" }, func(k8s.PodInfo) float64 { return 0 }, func(p k8s.PodInfo) bool { return p.Phase == "Pending" }); err != nil {
				return nil, err
			}
		case "pod_failed":
			if err := appendPodDetails(rule, func(k8s.PodInfo) string { return "Phase Failed" }, func(k8s.PodInfo) float64 { return 0 }, func(p k8s.PodInfo) bool { return p.Phase == "Failed" }); err != nil {
				return nil, err
			}
		case "pod_unknown":
			if err := appendPodDetails(rule, func(k8s.PodInfo) string { return "Phase Unknown" }, func(k8s.PodInfo) float64 { return 0 }, func(p k8s.PodInfo) bool { return p.Phase == "Unknown" }); err != nil {
				return nil, err
			}
		case "pod_oomkilled":
			if err := appendPodDetails(rule, func(p k8s.PodInfo) string { return fmt.Sprintf("OOMKilled=%t", p.OOMKilled) }, func(k8s.PodInfo) float64 { return 0 }, func(p k8s.PodInfo) bool { return p.OOMKilled }); err != nil {
				return nil, err
			}
		case "pod_image_pull_backoff":
			if err := appendPodDetails(rule, func(p k8s.PodInfo) string { return fmt.Sprintf("ImagePullBackOff=%t", p.ImagePullBackOff) }, func(k8s.PodInfo) float64 { return 0 }, func(p k8s.PodInfo) bool { return p.ImagePullBackOff }); err != nil {
				return nil, err
			}
		case "pod_crash_loop":
			if err := appendPodDetails(rule, func(p k8s.PodInfo) string { return fmt.Sprintf("CrashLoop=%t", p.CrashLoop) }, func(k8s.PodInfo) float64 { return 0 }, func(p k8s.PodInfo) bool { return p.CrashLoop }); err != nil {
				return nil, err
			}
		case "pod_not_ready":
			if err := appendPodDetails(rule, func(p k8s.PodInfo) string { return fmt.Sprintf("Ready=%t", p.Ready) }, func(k8s.PodInfo) float64 { return 0 }, func(p k8s.PodInfo) bool { return p.Phase == "Running" && !p.Ready }); err != nil {
				return nil, err
			}
		case "pod_evicted":
			if err := appendPodDetails(rule, func(p k8s.PodInfo) string { return fmt.Sprintf("Evicted=%t", p.Evicted) }, func(k8s.PodInfo) float64 { return 0 }, func(p k8s.PodInfo) bool { return p.Evicted }); err != nil {
				return nil, err
			}
		case "container_not_ready":
			if err := appendPodDetails(rule, func(p k8s.PodInfo) string { return fmt.Sprintf("ContainerNotReady=%t", p.ContainerNotReady) }, func(k8s.PodInfo) float64 { return 0 }, func(p k8s.PodInfo) bool { return p.Phase == "Running" && p.ContainerNotReady }); err != nil {
				return nil, err
			}
		}
	}
	return details, nil
}

func normalizeRuleType(ruleType string) string {
	switch ruleType {
	case "cpu_high":
		return "node_cpu"
	case "mem_high":
		return "node_memory"
	default:
		return ruleType
	}
}

func alertsFromDetails(inspectionID uint, details []inspModel.InspInspectionDetail) ([]inspModel.InspAlert, []string) {
	now := time.Now()
	alerts := make([]inspModel.InspAlert, 0)
	samples := make([]string, 0, 5)
	for _, detail := range details {
		if !detail.IsAnomaly {
			continue
		}
		content := fmt.Sprintf("%s: %s", detail.ResourceName, detail.Message)
		alerts = append(alerts, inspModel.InspAlert{
			InspectionID: inspectionID,
			RuleType:     detail.RuleType,
			Level:        alertLevel(detail.RuleType),
			Status:       "open",
			Content:      content,
			CollectedAt:  now,
		})
		if len(samples) < cap(samples) {
			samples = append(samples, content)
		}
	}
	return alerts, samples
}

func alertLevel(ruleType string) string {
	switch normalizeRuleType(ruleType) {
	case "node_not_ready", "pod_failed", "pod_oomkilled", "pod_crash_loop", "pod_evicted":
		return "critical"
	default:
		return "warning"
	}
}

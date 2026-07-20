package k8s

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	metricsclient "k8s.io/metrics/pkg/client/clientset/versioned"
)

type clientInspector struct {
	cs      *kubernetes.Clientset
	metrics *metricsclient.Clientset
}

func newClientInspector(kubeconfigPath string) (*clientInspector, error) {
	if _, err := os.Stat(kubeconfigPath); err != nil {
		return nil, fmt.Errorf("kubeconfig: %w", err)
	}
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("parse kubeconfig: %w", err)
	}
	cfg.Timeout = 20 * time.Second
	cfg.QPS = 20
	cfg.Burst = 40

	cs, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("clientset: %w", err)
	}

	var mc *metricsclient.Clientset
	if m, err := metricsclient.NewForConfig(cfg); err == nil {
		mc = m
	}

	return &clientInspector{cs: cs, metrics: mc}, nil
}

func (c *clientInspector) TestConnection(ctx context.Context) error {
	_, err := c.cs.Discovery().ServerVersion()
	return err
}

func (c *clientInspector) GetVersion(ctx context.Context) (string, error) {
	v, err := c.cs.Discovery().ServerVersion()
	if err != nil {
		return "", err
	}
	return v.GitVersion, nil
}

func (c *clientInspector) ListNodes(ctx context.Context) ([]NodeInfo, error) {
	list, err := c.cs.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	usage := c.nodeUsageMap(ctx)

	out := make([]NodeInfo, 0, len(list.Items))
	for _, n := range list.Items {
		ready := isNodeReady(n)
		status := "NotReady"
		if ready {
			status = "Ready"
		}
		info := NodeInfo{
			Name:               n.Name,
			Status:             status,
			DiskPressure:       hasNodeCondition(n, corev1.NodeDiskPressure),
			MemoryPressure:     hasNodeCondition(n, corev1.NodeMemoryPressure),
			PIDPressure:        hasNodeCondition(n, corev1.NodePIDPressure),
			NetworkUnavailable: hasNodeCondition(n, corev1.NodeNetworkUnavailable),
			Unschedulable:      n.Spec.Unschedulable,
		}
		if u, ok := usage[n.Name]; ok {
			info.CPUUsage = u.cpu
			info.MemUsage = u.mem
		}
		out = append(out, info)
	}
	return out, nil
}

func (c *clientInspector) ListPods(ctx context.Context, namespace string) ([]PodInfo, error) {
	list, err := c.cs.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	out := make([]PodInfo, 0, len(list.Items))
	for _, p := range list.Items {
		info := PodInfo{
			Namespace: p.Namespace,
			Name:      p.Name,
			Phase:     string(p.Status.Phase),
			Ready:     isPodReady(p),
			Evicted:   p.Status.Reason == "Evicted" || stringsContainsFold(p.Status.Message, "evicted"),
		}
		for _, cs := range p.Status.ContainerStatuses {
			info.RestartCount += cs.RestartCount
			if !cs.Ready {
				info.ContainerNotReady = true
			}
			if cs.LastTerminationState.Terminated != nil && cs.LastTerminationState.Terminated.Reason == "OOMKilled" {
				info.OOMKilled = true
			}
			if cs.State.Waiting != nil {
				switch cs.State.Waiting.Reason {
				case "ImagePullBackOff", "ErrImagePull":
					info.ImagePullBackOff = true
				case "CrashLoopBackOff":
					info.CrashLoop = true
				}
			}
		}
		for _, cs := range p.Status.InitContainerStatuses {
			if !cs.Ready && cs.State.Terminated == nil {
				info.ContainerNotReady = true
			}
			if cs.State.Waiting != nil {
				switch cs.State.Waiting.Reason {
				case "ImagePullBackOff", "ErrImagePull":
					info.ImagePullBackOff = true
				case "CrashLoopBackOff":
					info.CrashLoop = true
				}
			}
		}
		out = append(out, info)
	}
	return out, nil
}

func (c *clientInspector) ListNamespaces(ctx context.Context) ([]NamespaceInfo, error) {
	nsList, err := c.cs.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	podList, err := c.cs.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	counts := map[string]int{}
	for _, p := range podList.Items {
		counts[p.Namespace]++
	}
	out := make([]NamespaceInfo, 0, len(nsList.Items))
	for _, ns := range nsList.Items {
		out = append(out, NamespaceInfo{Name: ns.Name, PodCount: counts[ns.Name]})
	}
	return out, nil
}

type nodeUsage struct {
	cpu float64
	mem float64
}

func (c *clientInspector) nodeUsageMap(ctx context.Context) map[string]nodeUsage {
	result := map[string]nodeUsage{}
	if c.metrics == nil {
		return result
	}
	mlist, err := c.metrics.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return result
	}
	nodes, err := c.cs.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return result
	}
	capCPU := map[string]int64{}
	capMem := map[string]int64{}
	for _, n := range nodes.Items {
		if q, ok := n.Status.Capacity[corev1.ResourceCPU]; ok {
			capCPU[n.Name] = q.MilliValue()
		}
		if q, ok := n.Status.Capacity[corev1.ResourceMemory]; ok {
			capMem[n.Name] = q.Value()
		}
	}
	for _, m := range mlist.Items {
		u := nodeUsage{}
		if cpu := m.Usage.Cpu(); cpu != nil && capCPU[m.Name] > 0 {
			u.cpu = float64(cpu.MilliValue()) / float64(capCPU[m.Name]) * 100
		}
		if mem := m.Usage.Memory(); mem != nil && capMem[m.Name] > 0 {
			u.mem = float64(mem.Value()) / float64(capMem[m.Name]) * 100
		}
		result[m.Name] = u
	}
	return result
}

func isNodeReady(n corev1.Node) bool {
	for _, cond := range n.Status.Conditions {
		if cond.Type == corev1.NodeReady {
			return cond.Status == corev1.ConditionTrue
		}
	}
	return false
}

func hasNodeCondition(n corev1.Node, t corev1.NodeConditionType) bool {
	for _, cond := range n.Status.Conditions {
		if cond.Type == t {
			return cond.Status == corev1.ConditionTrue
		}
	}
	return false
}

func isPodReady(p corev1.Pod) bool {
	for _, cond := range p.Status.Conditions {
		if cond.Type == corev1.PodReady {
			return cond.Status == corev1.ConditionTrue
		}
	}
	return false
}

func stringsContainsFold(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

package k8s

import (
	"context"
	"fmt"
)

type stubInspector struct {
	connected bool
}

func newStubInspector(connected bool) *stubInspector {
	return &stubInspector{connected: connected}
}

func (s *stubInspector) TestConnection(ctx context.Context) error {
	if !s.connected {
		return fmt.Errorf("cluster unreachable (stub mode)")
	}
	return nil
}

func (s *stubInspector) GetVersion(ctx context.Context) (string, error) {
	if !s.connected {
		return "", fmt.Errorf("cluster unreachable")
	}
	return "v1.28.0-stub", nil
}

func (s *stubInspector) ListNodes(ctx context.Context) ([]NodeInfo, error) {
	if !s.connected {
		return nil, fmt.Errorf("cluster unreachable")
	}
	return []NodeInfo{
		{Name: "node-1", Status: "Ready", CPUUsage: 45, MemUsage: 62},
		{Name: "node-2", Status: "Ready", CPUUsage: 88, MemUsage: 55, DiskPressure: true},
		{Name: "node-3", Status: "NotReady", CPUUsage: 10, MemUsage: 20, Unschedulable: true, NetworkUnavailable: true},
	}, nil
}

func (s *stubInspector) stubPods() []PodInfo {
	return []PodInfo{
		{Namespace: "default", Name: "nginx-abc", Phase: "Running", RestartCount: 0, Ready: true},
		{Namespace: "default", Name: "api-xyz", Phase: "Running", RestartCount: 5, Ready: true},
		{Namespace: "default", Name: "job-fail", Phase: "Failed", RestartCount: 1, Ready: false},
		{Namespace: "kube-system", Name: "pull-err", Phase: "Pending", RestartCount: 0, Ready: false, ImagePullBackOff: true, ContainerNotReady: true},
		{Namespace: "default", Name: "crash-me", Phase: "Running", RestartCount: 12, Ready: false, CrashLoop: true, ContainerNotReady: true},
		{Namespace: "default", Name: "oom-pod", Phase: "Running", RestartCount: 2, Ready: false, OOMKilled: true, ContainerNotReady: true},
		{Namespace: "default", Name: "evicted-1", Phase: "Failed", RestartCount: 0, Ready: false, Evicted: true},
	}
}

func (s *stubInspector) ListPods(ctx context.Context, namespace string) ([]PodInfo, error) {
	if !s.connected {
		return nil, fmt.Errorf("cluster unreachable")
	}
	all := s.stubPods()
	if namespace == "" {
		return all, nil
	}
	out := make([]PodInfo, 0)
	for _, p := range all {
		if p.Namespace == namespace {
			out = append(out, p)
		}
	}
	return out, nil
}

func (s *stubInspector) ListNamespaces(ctx context.Context) ([]NamespaceInfo, error) {
	if !s.connected {
		return nil, fmt.Errorf("cluster unreachable")
	}
	counts := map[string]int{}
	for _, p := range s.stubPods() {
		counts[p.Namespace]++
	}
	out := make([]NamespaceInfo, 0, len(counts))
	for name, count := range counts {
		out = append(out, NamespaceInfo{Name: name, PodCount: count})
	}
	return out, nil
}

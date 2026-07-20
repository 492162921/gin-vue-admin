package k8s

import (
	"context"
	"os"
)

// NodeInfo 节点巡检信息（与 client-go 解耦的 DTO）
type NodeInfo struct {
	Name               string
	Status             string
	CPUUsage           float64
	MemUsage           float64
	Unschedulable      bool
	DiskPressure       bool
	MemoryPressure     bool
	PIDPressure        bool
	NetworkUnavailable bool
}

// PodInfo Pod 巡检信息
type PodInfo struct {
	Namespace         string
	Name              string
	Phase             string
	RestartCount      int32
	Ready             bool
	OOMKilled         bool
	ImagePullBackOff  bool
	CrashLoop         bool
	Evicted           bool
	ContainerNotReady bool
}

// NamespaceInfo 命名空间巡检信息
type NamespaceInfo struct {
	Name     string
	PodCount int
}

// Inspector K8s 只读巡检接口
type Inspector interface {
	TestConnection(ctx context.Context) error
	GetVersion(ctx context.Context) (string, error)
	ListNodes(ctx context.Context) ([]NodeInfo, error)
	ListPods(ctx context.Context, namespace string) ([]PodInfo, error)
	ListNamespaces(ctx context.Context) ([]NamespaceInfo, error)
}

// NewInspector 优先使用真实 kubeconfig；forceStub、环境变量 K8S_FORCE_STUB=1 或路径无效时回退 Stub。
func NewInspector(kubeconfigPath string, forceStub bool) Inspector {
	if forceStub || os.Getenv("K8S_FORCE_STUB") == "1" {
		return newStubInspector(true)
	}
	client, err := newClientInspector(kubeconfigPath)
	if err != nil {
		return newStubInspector(true)
	}
	return client
}

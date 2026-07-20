package k8s

import (
	"context"
	"os"
	"testing"
)

func TestStubListPodsContainsAnomalies(t *testing.T) {
	ins := NewInspector("", true)
	pods, err := ins.ListPods(context.Background(), "")
	if err != nil || len(pods) == 0 {
		t.Fatal(err)
	}
	var hasBad bool
	for _, p := range pods {
		if p.Phase != "Running" || p.RestartCount > 0 || !p.Ready {
			hasBad = true
		}
	}
	if !hasBad {
		t.Fatal("stub should include anomalous pods")
	}
}

func TestForceStubIgnoresBadPath(t *testing.T) {
	ins := NewInspector("Z:\\not-exist\\kubeconfig", true)
	if err := ins.TestConnection(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestBadKubeconfigReturnsConnectionError(t *testing.T) {
	ins := NewInspector("Z:\\not-exist\\kubeconfig", false)
	if err := ins.TestConnection(context.Background()); err == nil {
		t.Fatal("invalid kubeconfig must not be reported as connected")
	}
}

func TestEnvForceStub(t *testing.T) {
	t.Setenv("K8S_FORCE_STUB", "1")
	ins := NewInspector("Z:\\not-exist\\kubeconfig", false)
	if err := ins.TestConnection(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestStubListNodes(t *testing.T) {
	ins := NewInspector("", true)
	nodes, err := ins.ListNodes(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) < 3 {
		t.Fatalf("stub should have at least 3 nodes, got %d", len(nodes))
	}
}

func TestStubListNamespaces(t *testing.T) {
	ins := NewInspector("", true)
	nss, err := ins.ListNamespaces(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(nss) == 0 {
		t.Fatal("stub should return namespaces")
	}
	for _, ns := range nss {
		if ns.Name == "" {
			t.Fatal("namespace name should not be empty")
		}
	}
}

func TestStubGetVersion(t *testing.T) {
	ins := NewInspector("", true)
	v, err := ins.GetVersion(context.Background())
	if err != nil || v == "" {
		t.Fatalf("GetVersion: err=%v version=%q", err, v)
	}
}

func TestEnvForceStubUnset(t *testing.T) {
	os.Unsetenv("K8S_FORCE_STUB")
	ins := NewInspector("Z:\\not-exist\\kubeconfig", false)
	if err := ins.TestConnection(context.Background()); err == nil {
		t.Fatal("invalid kubeconfig must not be reported as connected")
	}
}

### Task 2: K8s Inspector (Real + Stub) + Unit Tests

**Files:**
- Create: `server/service/inspection/k8s/inspector.go`, `client.go`, `stub.go`
- Create: `server/service/inspection/k8s/inspector_test.go`
- Modify: `server/go.mod` / `go.sum`（`k8s.io/client-go` 及 metrics 相关依赖，版本与 Go 1.24 兼容）

**Interfaces:**
- Produces:

```go
package k8s

type NodeInfo struct {
	Name, Status string
	CPUUsage, MemUsage float64 // 0-100；无 metrics 时为 0
	Unschedulable bool
	// Conditions 相关布尔：DiskPressure, MemoryPressure, PIDPressure, NetworkUnavailable
}

type PodInfo struct {
	Namespace, Name, Phase string
	RestartCount int32
	Ready bool
	// 派生标记：OOMKilled, ImagePullBackOff, CrashLoop, Evicted 等
}

type NamespaceInfo struct {
	Name string
	PodCount int
}

type Inspector interface {
	TestConnection(ctx context.Context) error
	GetVersion(ctx context.Context) (string, error)
	ListNodes(ctx context.Context) ([]NodeInfo, error)
	ListPods(ctx context.Context, namespace string) ([]PodInfo, error)
	ListNamespaces(ctx context.Context) ([]NamespaceInfo, error)
}

func NewInspector(kubeconfigPath string, forceStub bool) Inspector
```

- `NewInspector`：`forceStub==true` 或路径无效/解析失败 → `StubInspector`；否则 `ClientInspector`。
- Stub：固定 ≥3 节点、≥若干异常 Pod，便于演示。

- [ ] **Step 1: Write Stub behavior tests first**

```go
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
```

- [ ] **Step 2: Run — expect FAIL**

```bash
cd server && go test ./service/inspection/k8s/ -count=1
```

- [ ] **Step 3: Implement Stub + Client + NewInspector**

Port 逻辑自源项目：  
`D:\项目\曜石\20260521-vibecoding\k8s-inspection-platform\backend\internal\k8s\`  
改造点：增加 `ListNamespaces`；JSON/字段命名对齐本包；`forceStub` 来自参数而非仅环境变量（环境变量 `K8S_FORCE_STUB=1` 可在 `NewInspector` 内 OR 进 forceStub）。

Client：用 client-go rest config from kubeconfig file；metrics 不可用时 CPU/Mem 记 0，不报错。

- [ ] **Step 4: Run tests PASS；go mod tidy**

```bash
cd server && go get k8s.io/client-go@latest && go mod tidy && go test ./service/inspection/k8s/ -count=1
```

- [ ] **Step 5: Commit**

```bash
git add server/service/inspection/k8s server/go.mod server/go.sum
git commit -m "feat(inspection): add k8s inspector with stub and real client"
```

---

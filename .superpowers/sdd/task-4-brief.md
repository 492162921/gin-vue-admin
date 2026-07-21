### Task 4: Inspection Engine + Task Run + History

**Files:**
- Create: `server/service/inspection/engine.go`, `task.go`
- Create: `server/api/v1/inspection/task.go`
- Create: `server/service/inspection/engine_test.go`
- Modify: routers + `web` task page + api + menu/api/casbin 追加 task
- Port evaluation：源 `backend/internal/service/services.go` 中规则分支

**Interfaces:**
- Produces:

```go
// Engine.Run 执行一次巡检并落库；namespace 过滤：rule.Scope=="namespace" 时 ListPods(ctx, rule.Namespace)，否则 ListPods(ctx, "")
func (e *Engine) Run(ctx context.Context, taskID uint) (*InspInspection, error)

type TaskService struct{ /* ... */ }
func (s *TaskService) Create/Update/Delete/List/Get(...)
func (s *TaskService) RunNow(ctx context.Context, taskID uint) (*InspInspection, error)
func (s *TaskService) ListInspections(ctx context.Context, taskID uint, info request.PageInfo) (...)
```

- 防重入：`sync.Map` 或 DB 状态，同 `taskID` 已有 `running` 则跳过并返回明确错误/日志。
- HTTP：`/inspection/task` CRUD、`/inspection/task/list`、`/inspection/task/run`、`/inspection/task/inspections`

- [ ] **Step 1: Failing engine test with Stub**

```go
func TestEngineRunWithStubMarksAnomalies(t *testing.T) {
	// NewMemoryDB migrate all insp tables
	// seed cluster (ForceStub), rules enabled node_not_ready + pod_not_running, task
	// force GVA_CONFIG.Inspection.ForceStub = true
	// Run → AnomalyCount > 0, details persisted
}
```

```go
func TestEngineRespectsNamespaceScope(t *testing.T) {
	// 使用可注入 Inspector 的 Engine（构造函数接受 Inspector 工厂或接口）
	// mock/fake：ListPods("", all)；ListPods("ns-a", only a)
	// rule scope=namespace namespace=ns-a → 只对 ns-a 的 pod 产生 detail
}
```

为实现可测性，`Engine` 必须支持注入 `func(cluster InspCluster) k8s.Inspector`，默认实现调用 `k8s.NewInspector(path, forceStub)`。

- [ ] **Step 2: Implement engine + task service + API + UI**

前端 `task.vue`：创建（选集群、多选规则、cron）、立即执行、历史抽屉（anomalyCount/summary/details）。

- [ ] **Step 3: Tests PASS**

```bash
cd server && go test ./service/inspection/ -count=1
```

- [ ] **Step 4: Commit**

```bash
git commit -am "feat(inspection): inspection engine, task run, and history UI"
```

（提交前 `git add` 明确文件列表，避免带上无关 config 改动。）

---

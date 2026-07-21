### Task 3: Cluster + Rule Services/APIs + Seed Rules + Frontend Pages

**Files:**
- Create: `server/service/inspection/cluster.go`, `rule.go`
- Create: `server/api/v1/inspection/cluster.go`, `rule.go`
- Create: `server/router/inspection/cluster.go`, `rule.go`（或合并进 `inspection.go`）
- Create: `server/source/inspection/rules.go`（内置 20 条规则种子）
- Create: `server/source/inspection/menu.go`, `api.go`, `casbin.go`（本 Task 只挂集群/规则菜单；后续 Task 追加）
- Modify: `server/initialize/register_init.go` blank import `source/inspection`
- Create: `web/src/api/inspection.js`（cluster/rule 段）
- Create: `web/src/view/inspection/cluster.vue`, `rule.vue`
- Test: `server/service/inspection/rule_service_test.go`

**Interfaces:**
- Produces:
  - `ClusterService.Create(ctx, name, kubeconfigBytes) (*InspCluster, error)` — 落盘到 `GVA_CONFIG.Inspection.KubeConfigDir`，再 `NewInspector` 探测 version/nodeCount/status
  - `ClusterService.Refresh / List / Get / Update / Delete / ListNodes / ListNamespaces`
  - `RuleService.Create/Update/Delete/SetEnabled/List`
  - 删除规则时若被 `insp_task_rules` 引用 → 返回业务错误
  - HTTP 路径约定（与 Casbin 种子一致）：
    - `/inspection/cluster` CRUD、`/inspection/cluster/list`、`/inspection/cluster/nodes`、`/inspection/cluster/namespaces`、`/inspection/cluster/refresh`
    - `/inspection/rule` CRUD、`/inspection/rule/list`、`/inspection/rule/enabled`

内置 `rule_type` 列表（必须种子齐全）：  
`node_cpu`, `node_memory`, `pod_restart`, `pod_restart_high`, `pod_not_running`, `node_not_ready`, `node_disk_pressure`, `node_memory_pressure`, `node_pid_pressure`, `node_network_unavailable`, `node_unschedulable`, `pod_pending`, `pod_failed`, `pod_unknown`, `pod_oomkilled`, `pod_image_pull_backoff`, `pod_crash_loop`, `pod_not_ready`, `pod_evicted`, `container_not_ready`

- [ ] **Step 1: Failing service test — delete rule blocked when referenced**

```go
func TestDeleteRuleBlockedWhenBoundToTask(t *testing.T) {
	db := testutil.NewMemoryDB(t, &InspRule{}, &InspTask{}, &InspCluster{})
	_ = db
	// create rule + task association, call RuleService.Delete, expect error
}
```

- [ ] **Step 2: Implement services + APIs + routers + seed**

菜单父级：

```text
Path: inspection, Component: view/routerHolder.vue, Title: K8s 巡检
  cluster → view/inspection/cluster.vue
  rule → view/inspection/rule.vue
```

Casbin：给权威 `888`（超管）挂上本 Task 所有 API。  
种子 `RegisterInit` order 选在 system 菜单之后（参考 media 的 order）。

前端 `cluster.vue`：列表、上传 kubeconfig 接入、刷新、抽屉展示 nodes + namespaces。  
`rule.vue`：列表、编辑阈值/scope/namespace、启停。

- [ ] **Step 3: Tests + build**

```bash
cd server && go test ./service/inspection/ -count=1 && go build -o NUL .
```

- [ ] **Step 4: Commit**

```bash
git add server/service/inspection server/api/v1/inspection server/router/inspection server/source/inspection server/initialize/register_init.go web/src/api/inspection.js web/src/view/inspection
git commit -m "feat(inspection): cluster and rule CRUD with seeds and UI"
```

---

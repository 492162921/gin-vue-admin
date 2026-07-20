# K8s Inspection Domain Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在 gin-vue-admin 中以原生业务域 `inspection` 实现 K8s 多集群巡检闭环（集群/规则/任务/告警/报告/总览），并补齐 namespace 过滤、OpenAI 兼容 AI 报告降级、真数据 Dashboard、节点与命名空间展示。

**Architecture:** 按 GVA 分层 `Router → API → Service → Model` 新增 `inspection` 域；鉴权复用 JWT+Casbin+菜单；调度通过同步 `sys_timed_tasks`（method=`RunInspectionTask`）；K8s 访问放在 `service/inspection/k8s`（real + stub）；报告先模板后 AI。

**Tech Stack:** Go 1.24 / Gin / GORM / client-go / robfig cron（经 GVA TimedTask）/ Vue3 / Element Plus / ECharts / Pinia

**Spec:** `docs/superpowers/specs/2026-07-20-k8s-inspection-gva-design.md`  
**Port from:** `D:\项目\曜石\20260521-vibecoding\k8s-inspection-platform`

## Global Constraints

- 表前缀必须为 `insp_`；模型嵌入 `global.GVA_MODEL`；JSON 字段用 camelCase（与 GVA 一致），不要照搬源项目 snake_case JSON。
- 一期不做 `CreatedBy`/`DeptId` 行级数据权限。
- Service 层禁止依赖 `gin.Context`；透传 `context.Context`。
- 列表分页统一 `request.PageInfo` + `info.LimitOffset()`。
- 统一响应 `response.OkWithData` / `FailWithMessage`；Swagger `@Success` 落到具体类型。
- kubeconfig / AI api-key 不得写入日志或接口回包明文。
- 前端样式优先 UnoCSS 原子类；菜单 Component 必须以 `view/` 开头精确匹配 vue 路径。
- 只读巡检，禁止任何 K8s 写操作。
- 每完成一个 Task 单独 commit。

## File Map（将创建/修改）

```text
server/config/inspection.go                          # 新建配置段
server/config/config.go                              # 挂 Inspection
server/config.yaml                                   # yaml 段
server/model/inspection/*.go                         # 实体 + request DTO
server/service/inspection/enter.go
server/service/inspection/cluster.go
server/service/inspection/rule.go
server/service/inspection/task.go
server/service/inspection/engine.go
server/service/inspection/alert.go
server/service/inspection/report.go
server/service/inspection/dashboard.go
server/service/inspection/webhook.go
server/service/inspection/k8s/inspector.go
server/service/inspection/k8s/client.go
server/service/inspection/k8s/stub.go
server/api/v1/inspection/*.go
server/router/inspection/*.go
server/source/inspection/menu.go, api.go, casbin.go, rules.go
server/initialize/{router,gorm,ensure_tables,register_init,timer}.go
server/service/enter.go, api/v1/enter.go, router/enter.go
server/go.mod                                        # 增加 k8s client-go 依赖
web/src/api/inspection.js
web/src/view/inspection/{dashboard,cluster,rule,task,alert,report}.vue
```

---

### Task 1: Config + Models + Domain Wiring Skeleton

**Files:**
- Create: `server/config/inspection.go`
- Create: `server/model/inspection/cluster.go`, `rule.go`, `task.go`, `inspection.go`, `alert.go`, `report.go`
- Create: `server/model/inspection/request/inspection.go`
- Create: `server/service/inspection/enter.go`
- Create: `server/api/v1/inspection/enter.go`
- Create: `server/router/inspection/enter.go`
- Create: `server/api/v1/inspection/health_stub.go`（临时空 API 组占位，后续任务替换）
- Create: `server/router/inspection/inspection.go`（先挂一个 GET ping 或仅空 Init，后续扩展）
- Modify: `server/config/config.go`, `server/config.yaml`
- Modify: `server/service/enter.go`, `server/api/v1/enter.go`, `server/router/enter.go`
- Modify: `server/initialize/gorm.go`, `ensure_tables.go`, `router.go`
- Test: `server/model/inspection/table_name_test.go`

**Interfaces:**
- Produces:
  - `config.Inspection` with fields: `KubeConfigDir string`, `ForceStub bool`, `WebhookURL string`, `AIBaseURL string`, `AIAPIKey string`, `AIModel string`
  - Models: `InspCluster`, `InspRule`, `InspTask`, `InspTaskRule`, `InspInspection`, `InspInspectionDetail`, `InspAlert`, `InspReport`；各自 `TableName()` 返回 `insp_*`
  - `service.ServiceGroupApp.InspectionServiceGroup` 空结构可编译

- [ ] **Step 1: Write failing table-name test**

```go
// server/model/inspection/table_name_test.go
package inspection

import "testing"

func TestTableNames(t *testing.T) {
	cases := map[string]string{
		InspCluster{}.TableName():           "insp_clusters",
		InspRule{}.TableName():              "insp_rules",
		InspTask{}.TableName():              "insp_tasks",
		InspInspection{}.TableName():        "insp_inspections",
		InspInspectionDetail{}.TableName():  "insp_inspection_details",
		InspAlert{}.TableName():             "insp_alerts",
		InspReport{}.TableName():            "insp_reports",
	}
	for got, want := range cases {
		if got != want {
			t.Fatalf("got %q want %q", got, want)
		}
	}
}
```

- [ ] **Step 2: Run test — expect FAIL (types missing)**

```bash
cd server && go test ./model/inspection/ -count=1
```

Expected: 编译失败 / undefined types

- [ ] **Step 3: Add config**

```go
// server/config/inspection.go
package config

type Inspection struct {
	KubeConfigDir string `mapstructure:"kube-config-dir" json:"kubeConfigDir" yaml:"kube-config-dir"`
	ForceStub     bool   `mapstructure:"force-stub" json:"forceStub" yaml:"force-stub"`
	WebhookURL    string `mapstructure:"webhook-url" json:"webhookUrl" yaml:"webhook-url"`
	AIBaseURL     string `mapstructure:"ai-base-url" json:"aiBaseUrl" yaml:"ai-base-url"`
	AIAPIKey      string `mapstructure:"ai-api-key" json:"aiApiKey" yaml:"ai-api-key"`
	AIModel       string `mapstructure:"ai-model" json:"aiModel" yaml:"ai-model"`
}
```

在 `config.Server` 增加：`Inspection Inspection \`mapstructure:"inspection" json:"inspection" yaml:"inspection"\``

在 `config.yaml` 追加：

```yaml
inspection:
  kube-config-dir: uploads/kubeconfigs
  force-stub: false
  webhook-url: ""
  ai-base-url: ""
  ai-api-key: ""
  ai-model: ""
```

- [ ] **Step 4: Implement models (camelCase JSON, GVA_MODEL)**

字段对齐源实体，但 JSON 用 camelCase。关键字段：

```go
// cluster.go 示例
type InspCluster struct {
	global.GVA_MODEL
	Name           string `json:"name" gorm:"size:128;comment:集群名"`
	KubeconfigPath string `json:"-" gorm:"size:512;comment:kubeconfig路径"`
	Status         string `json:"status" gorm:"size:32;default:unknown"`
	K8sVersion     string `json:"k8sVersion" gorm:"size:64"`
	NodeCount      int    `json:"nodeCount" gorm:"default:0"`
}
func (InspCluster) TableName() string { return "insp_clusters" }
```

`InspTask` 与 `InspRule` 多对多：`many2many:insp_task_rules;`  
`InspTask` 增加 `TimedTaskID *uint \`json:"timedTaskId"\`` 用于关联 `sys_timed_tasks`。  
`InspReport.Source` 取值：`ai` | `template`。  
`InspAlert.Status`：`open` | `acknowledged` | `closed`。

Request DTO（`request/inspection.go`）：分页列表查询结构体（含 name/status 等筛选字段）+ Create/Update 结构体。

- [ ] **Step 5: Wire enter.go + AutoMigrate + empty router**

仿照 media：

```go
// service/enter.go 增加
InspectionServiceGroup inspection.ServiceGroup

// initialize/gorm.go + ensure_tables.go 注册全部 insp_* 模型
// initialize/router.go:
inspectionRouter := router.RouterGroupApp.Inspection
inspectionRouter.InitInspectionRouter(PrivateGroup)
```

`ServiceGroup` / `ApiGroup` / `RouterGroup` 先放空 struct 或占位 `ClusterService` 空类型，保证 `go build ./...` 通过。

- [ ] **Step 6: Re-run test + build**

```bash
cd server && go test ./model/inspection/ -count=1 && go build -o NUL .
```

Expected: PASS / build OK

- [ ] **Step 7: Commit**

```bash
git add server/config server/model/inspection server/service/inspection server/api/v1/inspection server/router/inspection server/service/enter.go server/api/v1/enter.go server/router/enter.go server/initialize/gorm.go server/initialize/ensure_tables.go server/initialize/router.go server/config.yaml
git commit -m "feat(inspection): add config, models, and domain wiring skeleton"
```

---

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

### Task 5: Timed Task Sync + Alerts + Webhook

**Files:**
- Modify: `server/service/inspection/task.go`（同步 `sys_timed_tasks`）
- Create: `server/service/inspection/alert.go`, `webhook.go`
- Create: `server/api/v1/inspection/alert.go`
- Modify: `server/initialize/timer.go` — `task.Register("RunInspectionTask", ...)`
- Create: `server/service/inspection/webhook_test.go`, `alert_service_test.go`
- Frontend: `alert.vue` + api + menu/casbin

**Interfaces:**
- Produces:

```go
// timer.go 注册
task.Register("RunInspectionTask", "执行 K8s 巡检任务", func(ctx context.Context, params json.RawMessage) error {
	var p struct{ TaskID uint `json:"taskId"` }
	if err := json.Unmarshal(params, &p); err != nil { return err }
	_, err := InspectionServiceGroupApp.Engine.Run(ctx, p.TaskID) // 或 TaskService.RunNow
	return err
})
```

- `TaskService.Create/Update`：当 `status=active` 且 cron 非空 → `TimedTaskService.CreateTimedTask` / 更新：
  - `Name`: `insp-task-{id}`
  - `Spec`: cron（明确是否 withSeconds；默认 false，与源 5 段 cron 一致）
  - `ExecutorType`: `method`
  - `MethodName`: `RunInspectionTask`
  - `Params`: `{"taskId": <id>}`
  - `Enabled`: status==active
  - 回写 `InspTask.TimedTaskID`
- Delete/暂停：`ToggleTimedTask(false)` 或删除对应 timed task
- 引擎产生异常后：写 `InspAlert`，调用 `NotifyWebhook(content)`
- Webhook：URL 含 `oapi.dingtalk.com` → `{"msgtype":"text","text":{"content":"【日志】【信息】..."}}`；否则通用 JSON；校验钉钉 `errcode==0`
- Alert API：list（status/level/ruleType 筛选）、patch status、delete、batch-delete

- [ ] **Step 1: Webhook unit test (net/http/httptest)**

```go
func TestDingTalkPayloadContainsKeywords(t *testing.T) {
	// httptest.NewServer: assert body contains keywords and msgtype=text
}
```

- [ ] **Step 2: Implement sync + alert + webhook + UI**

引擎 `Run` 成功后若 `AnomalyCount>0` 自动 webhook（URL 空则 skip）。

- [ ] **Step 3: Tests**

```bash
cd server && go test ./service/inspection/ -count=1
```

- [ ] **Step 4: Commit**

```bash
git commit -m "feat(inspection): sync timed tasks, alerts, and webhook notify"
```

---

### Task 6: Reports (Template + AI Fallback) + Push

**Files:**
- Create: `server/service/inspection/report.go`, `report_ai.go`
- Create: `server/api/v1/inspection/report.go`
- Create: `server/service/inspection/report_test.go`
- Frontend: `report.vue`

**Interfaces:**
- Produces:

```go
func (s *ReportService) Generate(ctx context.Context, inspectionID uint) (*InspReport, error)
// 1) 组装巡检摘要+details
// 2) 若 AIBaseURL 与 AIAPIKey 非空 → 调 OpenAI 兼容 POST {base}/chat/completions
// 3) 成功 source=ai；否则 template Markdown，source=template
func (s *ReportService) Push(ctx context.Context, reportID uint) error // digest≤200 走 webhook
```

Chat body 最小字段：`model`, `messages`（system+user）。超时建议 30s。错误只记日志，不暴露 api-key。

定时巡检结束后可自动 `Generate`（与源项目一致）：在 `RunInspectionTask` 成功路径调用。

- [ ] **Step 1: Test template fallback when AI empty**

```go
func TestGenerateFallsBackToTemplate(t *testing.T) {
	// AI 配置为空，Generate → Source=="template", ContentMD 非空
}
```

- [ ] **Step 2: Optional test AI success with httptest mock**

```go
func TestGenerateUsesAIWhenConfigured(t *testing.T) {
	// httptest.NewServer 返回 OpenAI 兼容 choices[0].message.content
	// 配置指向该 server → Source=="ai"
}
```

- [ ] **Step 3: Implement + UI（列表/详情 Markdown/生成/推送/删除）**

- [ ] **Step 4: Commit**

```bash
git commit -m "feat(inspection): reports with AI and template fallback"
```

---

### Task 7: Dashboard APIs + Frontend Charts

**Files:**
- Create: `server/service/inspection/dashboard.go`
- Create: `server/api/v1/inspection/dashboard.go`
- Create: `server/service/inspection/dashboard_test.go`
- Create: `web/src/view/inspection/dashboard.vue`
- Modify: menu/api/casbin 增加「巡检总览」为分组下第一项

**Interfaces:**
- Produces HTTP：
  - `GET /inspection/dashboard/summary` → `{clusterCount, nodeCount, todayInspectionCount, openAlertCount}`
  - `GET /inspection/dashboard/resourceUsage` → 各集群或汇总 CPU/Mem（可来自最近一次巡检 detail 或实时 ListNodes；优先最近巡检聚合，无数据返回空数组）
  - `GET /inspection/dashboard/inspectionTrend` → 近 N 天每日 anomaly 合计（默认 7）
  - `GET /inspection/dashboard/alertDistribution` → 按 level / 按 cluster 计数

前端用项目已有 `echarts` / `vue-echarts`；禁止静态假数据。

- [ ] **Step 1: Dashboard summary test with seeded rows**

- [ ] **Step 2: Implement APIs + dashboard.vue**

- [ ] **Step 3: Commit**

```bash
git commit -m "feat(inspection): dashboard summary and real charts"
```

---

### Task 8: End-to-End Seed Polish + Manual Verification Checklist

**Files:**
- Modify: `server/source/inspection/menu.go` 确保菜单顺序：总览 → 集群 → 规则 → 任务 → 告警 → 报告
- Modify: 补齐所有 Casbin/API 种子遗漏
- Update: `docs/superpowers/specs/2026-07-20-k8s-inspection-gva-design.md` 状态改为「实现中/已落地」仅当本 Task 验证通过（或另开一行 Implementation status）
- Optional: `aiDoc/modules/inspection.md` 短说明（若仓库惯例需要；无则跳过）

- [ ] **Step 1: `go test ./service/inspection/... ./service/inspection/k8s/...` 全绿**

- [ ] **Step 2: `cd web && npm run build` 通过**

- [ ] **Step 3: 手工清单（Stub 模式）**

1. `inspection.force-stub: true` 重启后端  
2. 超管登录 → 侧栏见「K8s 巡检」  
3. 接入任意 kubeconfig 文本 → 集群 status 可用  
4. 创建任务绑规则 → 立即执行 → 历史有异常  
5. 告警列表有记录；配置 webhook（可用 https://webhook.site）可收到  
6. 生成报告（无 AI → template）；总览四图有数  

- [ ] **Step 4: Commit**

```bash
git commit -m "chore(inspection): finalize seeds and verification notes"
```

---

## Spec Coverage Self-Check

| Spec 项 | Task |
|:---|:---|
| 原生域 wiring / insp_ 表 | 1 |
| K8s real+stub / 只读 | 2 |
| 集群 CRUD + nodes/namespaces UI | 3 |
| 规则 CRUD + 20 种子 | 3 |
| namespace 过滤引擎 | 4 |
| 任务/立即执行/历史 | 4 |
| sys_timed_tasks 同步 | 5 |
| 告警 + Webhook | 5 |
| AI + template 报告 + push | 6 |
| Dashboard 四接口真数据 | 7 |
| 菜单顺序与联调 | 8 |
| 不做邮件/PDF/Landing/独立 RBAC | 全任务遵守 Global Constraints |

## Placeholder / Consistency Notes

- 所有 HTTP 路径统一前缀 `/inspection/...`，与菜单 API 种子一致。  
- `Engine` 注入 Inspector 工厂，保证 Task 4 单测不依赖真集群。  
- Timed task params 字段名固定 `taskId`（camelCase JSON）。  
- 前端 Component 路径：`view/inspection/dashboard.vue` 等与菜单种子逐字一致。

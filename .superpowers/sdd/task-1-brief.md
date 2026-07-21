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

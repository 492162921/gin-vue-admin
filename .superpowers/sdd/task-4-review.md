# Task 4 Re-Review — engine + task run + history

审查范围：`d0cfa8d1...124a418d`（含 `b94e50ac` 初版 + `124a418d` review fixes）。

## Spec verdict: 通过

- 通过：`NewEngine(InspectorFactory)` 可注入 Inspector；默认工厂调用 `k8s.NewInspector`。
- 通过：`rule.Scope == "namespace"` 时仅 `ListPods(ctx, rule.Namespace)`；`TestEngineRespectsNamespaceScope` 仅产出 `ns-a/bad-pod`。
- 通过：`sync.Map.LoadOrStore` 同 `taskID` 返回「正在执行」错误，`defer Delete` 释放占位。
- 通过：任务 CRUD、立即执行、按任务分页历史（含 Details 预加载）均已接通。
- 通过：`server/docs/docs.go`、`swagger.json`、`swagger.yaml` 已含 `/inspection/task`（GET/POST/PUT/DELETE）、`/inspection/task/list`、`/inspection/task/run`、`/inspection/task/inspections` 共七个接口。
- 通过：未引入 Task 5 定时同步、告警或 Webhook。

## Quality verdict: 通过

### Prior findings — closed

| ID | 原问题 | 状态 | 证据 |
|----|--------|------|------|
| P1 | Swagger 生成物缺失 `/inspection/task*` | **已关闭** | `124a418d` 重新 `swag init`；`swagger.json` 6098–6494 行含全部 task 路由 |
| P2 | 无并发防重入测试 | **已关闭** | `TestEngineRejectsConcurrentRunOnSameTask`：并发 Run 拒绝、结束后可再次 Run |

- 查询均通过 GORM 参数化 `Where`，未发现本任务新增 SQL 注入风险。
- 其余实现与初版 review 一致，无新增 P1/P2。

## Verification

```bash
cd server && go test ./service/inspection/ -count=1
# ok  github.com/flipped-aurora/gin-vue-admin/server/service/inspection  0.145s
```

## Approval

**Approved** — Spec 与 Quality 均满足 brief 与 Step 1–3 要求；P1/P2 已修复并有测试/文档证据。

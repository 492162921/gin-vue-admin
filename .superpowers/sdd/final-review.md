# K8s 巡检模块（feat/k8s-inspection）整体代码评审

**Base:** b1d2004e `→` **Head:** 34d76c2a（11 commits）
**范围:** server/{config,model,service,api,router,source}/inspection + web/src/{api,view}/inspection

## 总体结论

**Needs fixes（小修后可合并）** — 架构分层清晰，严格遵循 GVA Router→API→Service→Model 规范；`go build ./...`、`go test ./service/inspection/... ./model/inspection/...` 均通过；权限（JWT+MustChangePwd+Casbin+DataScope）、菜单、Casbin、API 种子逐一核对与 20 条内置规则齐全。存在 1 项前端功能性缺陷必须修复，另有若干中/低风险问题建议合并前处理。

## Critical（阻断）

无。

## Important（建议合并前修复）

1. **Dashboard 图表不渲染** — `web/src/view/inspection/dashboard.vue` 全部 4 个 `<Chart>` 用 `:option="..."`，但 `web/src/components/charts/index.vue` 声明的 prop 是 `options`（第16行），导致 ECharts 始终拿到空对象 `{}`，四张图（趋势/级别分布/资源使用/集群分布）实际不显示数据，与 spec「真数据图表」要求不符。属于功能性 bug，非样式问题。
2. **无效 kubeconfig 静默回退 Stub，误导集群可用性** — `service/inspection/k8s/inspector.go: NewInspector` 在 `newClientInspector` 解析失败（如格式错误的 kubeconfig）时直接返回 `stubInspector(connected:true)`；`ClusterService.probe` 因此把错误配置的集群标记为 `status=available` 并写入 3 个假节点数，用户会误认为真实集群健康。应区分“格式错误/不可解析”与“刻意 stub 模式”，前者应返回 `unreachable` 而不是伪造成功。
3. **告警级别/集群分布统计未过滤已关闭告警** — `service/inspection/dashboard.go: AlertDistribution` 两个查询都没有 `status != 'closed'`（或 `= 'open'`）过滤，long-running 后 Dashboard 的告警分布会被历史已处理告警持续污染，与「未关闭告警」类 KPI 语义不一致。

## Minor

4. `config/inspection.go: AIAPIKey` 带 `json:"aiApiKey"` 标签；当前未发现整体 config 回包的接口，无直接泄露路径，但与 Global Constraints「AI api-key 不得写入…接口回包明文」的防御性要求相悖，建议改 `json:"-"`。
5. `api/v1/inspection/health_stub.go` 为占位 API，含多余空行，功能上无害，收尾时可清理或直接删除（`/inspection/ping` 已无实际用途）。
6. Spec 中「实现路径已验证通过」的手工 Stub 联调清单（Task 8 Step 3 六项）未见执行记录/证据，仅有单元测试覆盖，建议合并前补跑一次手工验证或在 PR 描述中明确标注为"待手工验证"。
7. `ReportService.Push` 走 webhook 前未检查 `report.Digest` 是否为空/webhook 未配置以外的错误处理路径（如 4xx）只记录返回 error，前端目前也未展示失败详情，用户体验上可优化但非阻断。

## Spec 覆盖检查（对照设计文档 §4/§5）

| 项 | 状态 |
|---|---|
| 原生域 wiring，`insp_*` 表前缀，GVA_MODEL | ✅ |
| Casbin 路径+方法鉴权（888 全开），JWT/MustChangePwd/DataScope 链路 | ✅ |
| 集群 CRUD + nodes/namespaces + refresh | ✅（含 kubeconfig 落盘加密目录，回包 `json:"-"` 不回传路径） |
| 规则 CRUD + 20 条内置规则种子 + 启停 + 删除引用保护 | ✅ |
| 引擎 namespace 过滤（scope=namespace 时只拉指定 ns） | ✅（有专门单测覆盖） |
| 任务 CRUD + 立即执行 + 历史 + 防重入 | ✅（有并发单测覆盖） |
| sys_timed_tasks 同步（Create/Update/Delete/Toggle） | ✅ |
| 告警落库 + Webhook（钉钉 text / 通用 JSON） | ✅，但见 Important #3 |
| 报告 AI→模板降级 + 推送 | ✅，AI 超时/失败不阻断主流程 |
| Dashboard 四接口真数据 | ⚠️ 后端真实，前端渲染 bug（Important #1） |
| 菜单顺序：总览→集群→规则→任务→告警→报告 | ✅ |
| 不做写操作/邮件/PDF/独立RBAC | ✅ 未见违反 |

## 推荐手工 QA 步骤

1. `inspection.force-stub: true` 重启后端，超管登录确认侧栏「K8s 巡检」六个子菜单均可打开。
2. 接入一个**格式错误**的 kubeconfig 文本，确认集群状态应为 `unreachable`（当前会误报 `available`，验证 Important #2）。
3. 创建任务绑定「节点未就绪」+「Pod 未运行」规则并立即执行，检查历史异常数 >0，告警中心出现对应记录。
4. 打开巡检总览页，确认四张图表实际渲染出数据点（当前预期失败，验证 Important #1）。
5. 将一条告警状态置为 `closed` 后刷新总览，确认「告警级别/集群分布」饼图数值是否仍包含该已关闭告警（验证 Important #3）。
6. 配置 webhook.site 地址，触发一次异常巡检 + 手动推送报告，确认收到钉钉格式与通用 JSON 两种 payload。
7. 生成报告（AI 未配置）确认 `source=template`；填入任意 OpenAI 兼容 endpoint 后重新生成确认 `source=ai`。

## Fixes applied

- Dashboard 的四个 `<Chart>` 调用均改用 `options` prop，恢复趋势、告警级别、资源使用和集群分布图表渲染。
- 非 Stub 模式下，无效或不可解析的 kubeconfig 会返回连接/验证错误；创建、更新不会持久化该配置，刷新会保存 `unreachable` 状态并返回明确错误。仅 `inspection.force-stub` 或 `K8S_FORCE_STUB=1` 启用 Stub。
- 告警级别和集群分布查询均排除 `closed` 告警；摘要的 `openAlertCount` 原本已仅统计 `open`。
- `AIAPIKey` 已从 JSON 序列化中排除；`/inspection/ping` 已说明其仅检查路由可达性。

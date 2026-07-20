# Task 8 完成报告：巡检种子与验证收尾

提交：`chore(inspection): finalize seeds and verification notes`

## 实现

- 已核对 K8s 巡检菜单排序：总览（1）→ 集群（2）→ 规则（3）→ 任务（4）→ 告警（5）→ 报告（6）。
- 审计路由后补充 `/inspection/ping` 的 GET API 种子；Casbin 初始化复用同一份 `inspectionAPIs()`，因此同步获得超级管理员权限。
- `web/src/pathInfo.json` 已包含全部六个巡检视图映射，无需生成或修改。
- 设计说明状态已更新为 `feat/k8s-inspection` 已落地。

## 自动验证

- `cd server && go test ./service/inspection/... ./service/inspection/k8s/... -count=1`：通过。
- `cd web && npm run build`：通过。

## 手工验证清单（Stub 模式，待执行）

- [ ] 设置 `inspection.force-stub: true` 并重启后端。
- [ ] 超级管理员登录后，确认侧栏显示「K8s 巡检」及六项菜单顺序。
- [ ] 提交任意 kubeconfig 文本，确认集群状态可用并可查看节点、命名空间。
- [ ] 创建关联规则的任务并立即执行，确认历史含异常记录。
- [ ] 确认告警列表有记录；配置 webhook.site 后确认收到 webhook。
- [ ] 生成报告（未配置 AI 时为模板）；确认总览四张图有数据。

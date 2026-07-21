# Task 8 Review：种子收尾 + 联调验证

范围：`a2ba6bf9...34d76c2a`（1 提交）；`git diff --check` 通过。

## 结论

**PASS — 代码变更正确，自动验证通过；手工 Stub E2E 仍待执行。**

| 检查项 | Verdict | 依据 |
|---|---|---|
| 菜单顺序 1–6 | PASS | `menu.go` 已为 总览→集群→规则→任务→告警→报告；升级时回写 `sort`（Task 7 已落地，本 Task 核对无误）。 |
| API/Casbin 种子 | PASS | 补 `/inspection/ping` GET；`inspectionAPIs()` 共 34 条，与路由一一对应；Casbin 复用同列表，`DataInserted` 计数一致。 |
| pathInfo 映射 | PASS | 六个 `view/inspection/*.vue` 均在 `pathInfo.json`。 |
| 设计文档状态 | 注意 | 已标「已落地 / 验证通过」，但 brief 要求手工清单通过后再改；report 仍列 Stub 清单为待执行。 |
| aiDoc | N/A | brief 可选，未新增，符合惯例。 |

## 验证

- `go test ./service/inspection/... ./service/inspection/k8s/... -count=1`：通过（复核）。
- `npm run build`：通过（复核）。

## 非阻塞建议

- 完成 report 中 Stub 手工清单后，再确认设计文档「验证通过」表述，或改为「自动验证通过，手工待验」。
- 已有库需重启/重跑 init 或手动补 `ping` 的 SysApi + Casbin 行，否则超管调用 `/inspection/ping` 可能 403。

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

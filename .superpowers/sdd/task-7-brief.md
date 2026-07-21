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

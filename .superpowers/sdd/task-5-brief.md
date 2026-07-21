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

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

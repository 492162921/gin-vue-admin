# Task 6 完成报告：巡检报告

提交：`f7eb0b29 feat(inspection): reports with AI and template fallback`

## 实现

- 新增报告生成、列表、详情、推送和删除接口；模板报告为默认降级路径。
- 配置 AI 后通过 OpenAI 兼容 `/chat/completions` 生成 Markdown；HTTP 客户端超时为 30 秒，日志不记录 API Key。
- 报告摘要截断至 200 个字符后推送 Webhook；定时巡检成功后自动生成报告。
- 新增报告菜单、Casbin/API 种子、Vue 报告页与 Swagger 文档。

## 验证

- `go test ./service/inspection/... -count=1`
- `go build -o NUL .`
- `npm run build`
- `go run github.com/swaggo/swag/cmd/swag@v1.16.4 init -g main.go -o docs`

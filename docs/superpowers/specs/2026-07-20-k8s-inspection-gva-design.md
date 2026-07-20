# K8s 巡检能力迁入 gin-vue-admin — 设计说明

| 项 | 内容 |
|:---|:---|
| 日期 | 2026-07-20 |
| 状态 | 已评审（待实现计划） |
| 源项目 | `D:\项目\曜石\20260521-vibecoding\k8s-inspection-platform` |
| 目标仓库 | `D:\GoProjects\gin-vue-admin` |

## 1. 目标与范围

将源项目「AI 驱动 Kubernetes 资源巡检平台」的能力，以 **GVA 原生业务模块** 方式实现于本仓库，而不是独立第二套后端或插件。

### 1.1 范围（方案 B）

**纳入：**

- 源项目已实现的闭环：登录复用 GVA → 集群（kubeconfig）→ 规则 → 定时/手动巡检 → 告警 → 模板报告 → 通用 Webhook（含钉钉文本）
- 补齐缺口：
  - 规则 `scope=namespace` 时引擎真正按命名空间过滤
  - OpenAI 兼容大模型生成报告，失败/未配置降级模板
  - Dashboard 真数据图表（summary / resource-usage / inspection-trend / alert-distribution）
  - 前端展示集群节点列表与命名空间概览（对接 nodes / namespaces API）

**不纳入：**

- 源项目独立 RBAC / 用户角色管理页（改用 GVA Casbin + 菜单）
- 邮件告警、PDF 导出、独立 Landing 页
- 对 K8s 的写操作（重启 Pod、扩缩容等）
- 进程外保留原 backend 做代理

### 1.2 已确认决策

| 决策点 | 选择 |
|:---|:---|
| 集成方式 | 原生业务域 `inspection`，非 GVA 插件 |
| 权限 | 完全复用 GVA JWT + Casbin + 动态菜单 |
| AI | OpenAI 兼容 `base_url` / `api_key` / `model`；失败降级模板 |
| UI | 融入 GVA 后台风格；侧栏独立分组「K8s 巡检」 |
| Dashboard | 独立菜单「巡检总览」，不替换 GVA 默认首页 |
| 通知 | 仅通用 Webhook（含钉钉 text），不做邮件 |
| 实现路径 | 按 GVA 分层一次落地（Router → API → Service → Model） |

## 2. 架构与模块边界

### 2.1 后端目录

| 路径 | 职责 |
|:---|:---|
| `server/model/inspection/` | 实体与 request/response DTO |
| `server/service/inspection/` | 集群、规则、任务、巡检引擎、告警、报告、Webhook |
| `server/api/v1/inspection/` | HTTP 入参与响应，不写业务逻辑 |
| `server/router/inspection/` | 挂到 `PrivateGroup`（JWT → MustChangePwd → Casbin → DataScope） |
| `server/service/inspection/k8s/` | client-go + Stub（仅本域使用，不放全局 utils） |
| `server/source/inspection/` | 菜单、API、Casbin、内置规则种子 |

在 `router/enter.go`、`api/v1/enter.go`、`service/enter.go`、`initialize/router.go`、`initialize/gorm.go` 中注册该域，风格对齐 `example` / `media`。

### 2.2 前端目录

| 路径 | 职责 |
|:---|:---|
| `web/src/view/inspection/` | 总览、集群、规则、任务、告警、报告 |
| `web/src/api/inspection*.js` | 对接后端 |
| 菜单种子 | 分组「K8s 巡检」：总览 → 集群 → 规则 → 任务 → 告警 → 报告 |

### 2.3 复用与不碰

- **复用：** 登录、用户、角色、Casbin、动态菜单、统一响应 `{code,data,msg}`、分页 `PageInfo`、`sys_timed_tasks` 调度基础设施
- **不碰：** 源项目自建 auth；独立 Landing；第二套 cron 进程（调度通过同步 GVA 定时任务完成）

### 2.4 运行时数据流

```text
浏览器 → GVA 鉴权中间件 → inspection API
                              ↓
                    service/inspection
                    ├─ Cluster / Rule / Task
                    ├─ Engine (K8s real | stub)
                    ├─ Alert + Webhook
                    └─ Report (AI → fallback template)
                              ↓
                         GORM / 默认库（多为 MySQL）
```

## 3. 数据模型

表前缀 `insp_`，避免与系统表冲突。一期 **不做** 行级数据权限字段（`CreatedBy`/`DeptId`）；访问控制仅靠菜单 + Casbin。

| 表 | 说明 |
|:---|:---|
| `insp_clusters` | 名称、kubeconfig 路径或加密存储引用、状态、版本、节点数 |
| `insp_rules` | `rule_type`、阈值、`scope`（cluster/namespace）、`namespace`、启用 |
| `insp_tasks` | 集群 ID、cron、调度类型、状态 |
| `insp_task_rules` | 任务-规则多对多 |
| `insp_inspections` | 任务 ID、状态、异常数、摘要、起止时间 |
| `insp_inspection_details` | 规则类型、资源标识、是否异常、指标值、消息 |
| `insp_alerts` | 巡检 ID、级别、状态（open / acknowledged / closed）、内容 |
| `insp_reports` | 巡检 ID、标题、Markdown、digest、`source`（`ai` \| `template`） |

**种子：**

- 约 20 条内置规则（与源项目 `builtinRules` 对齐）
- 侧栏菜单 + 对应 API 的 Casbin 策略
- 可选预置角色「巡检运维」「巡检只读」；超级管理员默认全开

**kubeconfig：** 仅服务端落盘或加密存储；列表/详情接口不回传明文全文。

## 4. API 与页面

统一前缀使用 GVA `RouterPrefix`；Swagger 注释与真实行为一致。

### 4.1 API（私有）

| 模块 | 能力 |
|:---|:---|
| 集群 | CRUD、`GET :id/nodes`、`GET :id/namespaces`、`POST :id/refresh` |
| 规则 | CRUD、启停 |
| 任务 | CRUD、`POST :id/run`、`GET :id/inspections`；启停时同步 `sys_timed_tasks` |
| 告警 | 筛选列表、改状态、删除 / 批量删除 |
| 报告 | 列表、详情、`POST generate`、`POST :id/push`、删除 |
| 总览 | `summary`、`resource-usage`、`inspection-trend`、`alert-distribution` |

权限以 **Casbin 路径 + HTTP 方法** 为准，不再使用源项目字符串权限点（如 `cluster:read`）。

### 4.2 前端页面

| 菜单 | 能力 |
|:---|:---|
| 巡检总览 | 4 KPI + 资源图 + 巡检趋势 + 告警分布（ECharts，真数据） |
| 集群管理 | 接入/编辑/删除/刷新；详情或抽屉展示节点 |
| 巡检规则 | 编辑阈值与作用域、启停；规则类型枚举 |
| 巡检任务 | 绑集群与规则、Cron、立即执行、历史 |
| 告警中心 | 筛选、确认/关闭、批量删除 |
| 巡检报告 | 按巡检生成、Markdown 查看、Webhook 推送、删除 |

风格：GVA 表格 + 抽屉/对话框；可用 `v-auth` 与菜单 API 对齐。

## 5. 巡检引擎、调度、AI、Webhook

### 5.1 引擎流程

1. 立即执行或定时触发  
2. 加载集群与启用规则；`scope=namespace` 时只拉指定命名空间  
3. Inspector 拉取节点/Pod（及 metrics，若可用）  
4. 按规则判定 → 写 inspection + details  
5. 异常 → alerts；配置了 Webhook 则异步推送  
6. 定时完成后可自动生成报告；报告页支持手动生成  

内置规则类型与源项目一致（节点 CPU/内存、NotReady、压力条件、Pod 重启/非 Running/CrashLoop/OOM 等）。

### 5.2 K8s 访问

- 默认：client-go + 真实 kubeconfig  
- `K8S_FORCE_STUB=1` 或配置无效：回退 Stub 样例数据（演示用）  
- 只读

### 5.3 调度

- 巡检任务 CRUD/启停时同步写入/更新/禁用 GVA `sys_timed_tasks`，由现有 `LoadTimedTasks` / TimedTask 服务调度；执行入口调用 `InspectionService.Run`  
- 同一任务防重入：上次未结束则跳过并记日志  

### 5.4 AI 报告

- 配置项：`ai.base-url`、`ai.api-key`、`ai.model`（OpenAI 兼容 chat completions）  
- 成功：`source=ai`  
- 未配置 / 超时 / 非 2xx：模板 Markdown，`source=template`，不阻断巡检主流程  

### 5.5 Webhook

- 单一 URL；钉钉 text 与通用 JSON（对齐源项目）  
- 未配置：跳过推送，仅落库  
- 用途：巡检异常自动推；报告手动推摘要  

### 5.6 错误处理

- 连通失败：集群状态异常，当次 inspection 标记 failed，调度器继续  
- 单规则异常：记 detail/日志，尽量继续其余规则  
- 日志不输出 kubeconfig / API Key 明文  

## 6. 配置

在 `server/config.yaml`（及对应 `config` 结构体）增加 inspection 相关段，至少包括：

- kubeconfig 存储目录  
- `K8S_FORCE_STUB`（或等价配置键）  
- Webhook URL  
- AI `base-url` / `api-key` / `model`  

敏感项不进前端构建产物；不提交真实密钥。

## 7. 验证策略

- 后端：规则判定与 Stub 路径单元测试；引擎关键路径可测  
- 联调：无真实集群时用 Stub 跑通「任务执行 → 告警 → 报告 → Webhook（可 mock）」  
- 前端：各页 CRUD + 总览四图与后端数据一致  
- 有真实集群时再验 refresh/nodes 与非 Stub 巡检  

## 8. 实现分期建议（供后续 plan 拆分）

实现计划阶段建议按可交付增量拆分，例如：

1. 模型 + AutoMigrate + 菜单/Casbin 种子 + 空壳路由页面  
2. 集群 + K8s client/Stub + 节点展示  
3. 规则 + 任务 + 引擎 + 立即执行/历史  
4. 调度同步 + 告警 + Webhook  
5. 报告（模板 → AI）+ Dashboard 四接口与图表  

每期保持可运行、可演示。

## 9. 非目标回顾

- 不移植源项目独立用户体系  
- 不做邮件 / PDF / Landing  
- 不做 K8s 写操作  
- 不做行级数据权限（一期）  
- 不以插件或外部进程方式交付主功能  

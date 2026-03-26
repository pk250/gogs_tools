---
story_id: 1-7
story_key: NOKEY-1-7
title: Dashboard 提交列表与任务详情页
epic: epic-1
status: done
created: '2026-03-25'
---

# Story 1-7：Dashboard 提交列表与任务详情页

## 目标

提供 Dashboard 提交列表页和任务详情页，让开发者可以查看历史编译记录、筛选任务、查看详细日志和下载产物。

---

## 依赖

| 前置 Story | 依赖内容 |
|------------|----------|
| Story 1-1 | `BuildTask` 模型、`ReviewResult` 模型 |
| Story 1-4 | 任务状态枚举（pending/running/success/failed）|
| Story 1-5 | 产物文件路径约定（`data/artifacts/:taskId/`）|
| Story 1-6 | WebSocket 日志推送（详情页实时日志区域）|

---

## 已实现代码上下文

### Dashboard 列表页

**文件：** `controllers/dashboard.go`

- `GET /dashboard` — 分页列表，pageSize=20，支持按 `repo`/`status`/`author` 过滤
- 统计今日 success/failed/running 数量
- 为每条任务计算 `ReviewResult` 聚合状态（通过/有警告/有错误）
- 模板：`views/dashboard/index.tpl`

```go
qs := o.QueryTable("build_task")
if filterRepo != "" {
    qs = qs.Filter("RepoName", filterRepo)
}
// ...
taskReviewStatus := make(map[int64]reviewSummary)
for _, t := range tasks {
    var results []models.ReviewResult
    o.QueryTable("review_result").Filter("TaskId", t.Id).All(&results)
    // 聚合：有 fail → danger，有 warn → warning，其余 → success
}
```

### 任务详情页

**文件：** `controllers/build.go`

- `GET /build/detail/:taskId` — 读取任务、ReviewResult 列表、日志内容
- statusClass 映射（pending/running/success/failed → Bootstrap 颜色类）
- 各 ReviewResult 按 `CheckType` 分组展示
- 模板：`views/build/detail.tpl`

### 产物下载

**文件：** `controllers/build.go`

```go
// GET /api/build/:taskId/download/:filename
func (this *BuildController) Download() {
    path := filepath.Join(".", "data", "artifacts", taskIdStr, filename)
    this.Ctx.Output.Download(path, filename)
}
```

### 手动触发

**文件：** `controllers/build.go`

```go
// POST /api/build/trigger
func (this *BuildController) Trigger() {
    // 解析 JSON body，调用 services.Enqueue()
    // 队列满返回 429
}
```

### 路由

**文件：** `routers/router.go`

```go
beego.Router("/dashboard", &controllers.DashboardController{}, "get:Index")
beego.Router("/build/detail/:taskId", &controllers.BuildController{}, "get:Detail")
beego.Router("/api/build/trigger", &controllers.BuildController{}, "post:Trigger")
beego.Router("/api/build/:taskId/download/:filename", &controllers.BuildController{}, "get:Download")
```

---

## 验证清单

- [x] `GET /dashboard` 返回分页任务列表
- [x] 筛选参数（repo/status/author）正常过滤
- [x] 今日统计数字正确显示
- [x] 任务列表显示「审查状态」列（有 ReviewResult 时）
- [x] `GET /build/detail/:taskId` 显示任务详情
- [x] 详情页日志区通过 WebSocket 实时推送
- [x] 产物下载链接有效
- [x] 手动触发接口返回 task_id

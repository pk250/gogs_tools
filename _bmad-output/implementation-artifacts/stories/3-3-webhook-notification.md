---
story_id: 3-3
story_key: NOKEY-3-3
title: 外部 Webhook 回调通知
epic: epic-3
status: done
created: '2026-03-25'
---

# Story 3-3：外部 Webhook 回调通知

## 目标

编译完成后 POST 回调外部系统，携带任务状态和产物下载链接，回调失败时记录日志并支持手动重试，与其他系统集成。

---

## 依赖

| 前置 Story | 依赖内容 |
|------------|----------|
| Story 1-1 | `RepoConfig.WebhookEnabled`、`RepoConfig.WebhookUrl` 字段 |
| Story 1-3 | 仓库配置页（Webhook URL 设置入口）|
| Story 1-5 | 编译完成后回调点 |
| Story 1-7 | 产物下载链接生成 |

---

## 已实现代码上下文

### Webhook 服务

**文件：** `services/notifier/webhook.go`

- `SendWebhook(task BuildTask)` — 向仓库配置的 `webhook_url` 发送 POST 请求
- 请求体（JSON）：
  ```json
  {
    "task_id": 123,
    "repo_name": "my-repo",
    "status": "success",
    "commit_hash": "abc1234",
    "duration_seconds": 42,
    "artifact_url": "/api/build/123/download/firmware.hex"
  }
  ```
- 不包含产物文件本体，仅提供下载链接
- 回调失败时记录日志（`logs/webhook.log`）

```go
// services/notifier/webhook.go
func SendWebhook(task models.BuildTask) {
    cfg := getRepoConfig(task.RepoName)
    if !cfg.WebhookEnabled || cfg.WebhookUrl == "" {
        return
    }
    payload := buildPayload(task)
    resp, err := http.Post(cfg.WebhookUrl, "application/json", payload)
    if err != nil || resp.StatusCode >= 400 {
        beego.Error("webhook failed:", err)
    }
}
```

### 手动重试接口

**文件：** `controllers/build.go`

```go
// POST /api/build/:taskId/webhook-retry
func (this *BuildController) WebhookRetry() {
    // 读取任务，异步调用 notifier.SendWebhook(task)
    this.Data["json"] = map[string]interface{}{"code": 0, "message": "webhook 回调已触发"}
}
```

### 仓库配置

**文件：** `controllers/repo.go` / `views/repo/config.tpl`

- 仓库配置页支持设置 Webhook URL 和启用开关

---

## 验证清单

- [x] 仓库配置支持设置外部 Webhook URL
- [x] 编译完成后自动 POST 回调
- [x] 回调体包含：状态、commit hash、编译耗时、产物下载链接
- [x] 不包含产物文件本体
- [x] 回调失败时记录日志
- [x] 支持通过 `/api/build/:taskId/webhook-retry` 手动重试

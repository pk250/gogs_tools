---
story_id: 1-8
story_key: NOKEY-1-8
title: 邮件通知
epic: epic-1
status: done
created: '2026-03-25'
---

# Story 1-8：邮件通知

## 目标

编译任务完成后（成功或失败）自动发送邮件通知，告知相关人员编译结果，并在 Epic 2 完成后包含审查报告摘要。

---

## 依赖

| 前置 Story | 依赖内容 |
|------------|----------|
| Story 1-1 | `BuildTask` 模型（`NotifyEmails` 字段）|
| Story 1-5 | 编译完成后回调点 |

---

## 已实现代码上下文

### SMTP 通知服务

**文件：** `services/notifier/smtp.go`

- 从 beego 配置读取 SMTP 参数（`smtp_host`, `smtp_port`, `smtp_user`, `smtp_pass`, `smtp_from`）
- 编译完成后调用 `SendBuildResult(task BuildTask, results []ReviewResult)`
- 邮件内容包含：任务状态、仓库名、commit hash、编译耗时
- Epic 2 完成后升级：附带各审查项目一行摘要

```go
// services/notifier/smtp.go
func SendBuildResult(task models.BuildTask, results []models.ReviewResult) {
    // 读取 smtp 配置
    // 组装 HTML 邮件正文
    // 发送至 task.NotifyEmails（逗号分隔）
}
```

### 调用点

**文件：** `services/worker.go`（编译完成后）

```go
go notifier.SendBuildResult(task, reviewResults)
```

### Webhook 回调（配套）

**文件：** `services/notifier/webhook.go`

- POST 回调外部 URL，携带任务状态、commit hash、耗时、产物下载链接
- 失败时记录日志，支持手动重试（`/api/build/:taskId/webhook-retry`）

---

## 验证清单

- [x] 编译成功/失败后自动发送邮件
- [x] 邮件收件人来自仓库配置的 `notify_emails` 字段
- [x] 邮件包含任务状态、仓库名、commit hash
- [x] SMTP 配置错误时记录日志，不影响主流程
- [x] Webhook 回调在编译完成后触发
- [x] Webhook 失败可通过接口手动重试

---
story_id: 2-4
story_key: NOKEY-2-4
title: 聚合审查报告展示
epic: epic-2
status: done
created: '2026-03-25'
---

# Story 2-4：聚合审查报告展示

## 目标

在任务详情页将编译、PC-Lint、AI审查、Git规范四个审查结果聚合展示，并升级邮件通知包含审查摘要，Dashboard 增加「审查状态」列。

---

## 依赖

| 前置 Story | 依赖内容 |
|------------|----------|
| Story 1-7 | Dashboard 列表页、任务详情页框架 |
| Story 1-8 | 邮件通知服务 |
| Story 2-1 | lint ReviewResult |
| Story 2-2 | git ReviewResult |
| Story 2-3 | ai ReviewResult |

---

## 已实现代码上下文

### 任务详情页阶段进度条

**文件：** `views/build/detail.tpl`

- 阶段进度条：编译 → PC-Lint → AI审查 → Git规范 → 通知
- 四个报告区块各自独立折叠面板
- 失败/有问题的面板自动展开
- 各面板通过 `check_type` 字段区分（compile/lint/ai/git）

### Dashboard 审查状态列

**文件：** `controllers/dashboard.go`

```go
// 为每个任务聚合 ReviewResult
taskReviewStatus := make(map[int64]reviewSummary)
for _, t := range tasks {
    var results []models.ReviewResult
    o.QueryTable("review_result").Filter("TaskId", t.Id).All(&results)
    // 有 fail → 有错误(danger)，有 warn → 有警告(warning)，其余 → 通过(success)
}
this.Data["taskReviewStatus"] = taskReviewStatus
```

**文件：** `views/dashboard/index.tpl`

- 任务列表增加「审查状态」列
- 无 ReviewResult 的任务不显示该列（编译中或 Epic 2 前的历史任务）

### 邮件通知升级

**文件：** `services/notifier/smtp.go`

- `SendBuildResult(task, results []ReviewResult)` 接收审查结果列表
- 邮件正文增加各审查项目一行摘要：
  - 编译：成功/失败
  - PC-Lint：X 警告 / X 错误
  - AI审查：通过/有建议
  - Git规范：合规/不合规

---

## 验证清单

- [x] 任务详情页显示四阶段进度条
- [x] 四个报告区块各自独立折叠
- [x] 失败/有问题的面板自动展开
- [x] Dashboard 列表显示「审查状态」列
- [x] 审查状态正确区分：通过/有警告/有错误
- [x] 邮件通知包含审查报告摘要

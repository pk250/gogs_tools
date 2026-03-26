---
story_id: 2-3
story_key: NOKEY-2-3
title: AI 代码审查集成
epic: epic-2
status: done
created: '2026-03-25'
---

# Story 2-3：AI 代码审查集成

## 目标

编译完成后调用云端 AI（Claude/OpenAI）对代码变更 diff 进行语义审查，结果保存至数据库并在任务详情页展示。

---

## 依赖

| 前置 Story | 依赖内容 |
|------------|----------|
| Story 1-1 | `ReviewResult` 模型（`check_type`=`ai`）|
| Story 1-5 | 编译完成后回调点、git diff 获取 |

---

## 已实现代码上下文

### AI 审查接口（适配器模式）

**文件：** `services/ai/interface.go`

```go
type AIReviewer interface {
    Review(diff string, prompt string) (string, error)
}
```

### Claude 适配器

**文件：** `services/ai/claude.go`

- 实现 `AIReviewer` 接口，调用 Anthropic API
- 使用管理员配置的 API Key（AES-256-GCM 加密存储，界面脱敏显示 `****`）

### OpenAI 适配器

**文件：** `services/ai/openai.go`

- 实现 `AIReviewer` 接口，调用 OpenAI API
- 同样支持加密存储 API Key

### AI 审查调度

**文件：** `services/ai_reviewer.go`

- `RunAIReview(task BuildTask) error`
- 根据管理员配置选择服务商（claude/openai）
- 获取本次提交 diff（git diff HEAD~1..HEAD）
- 调用对应适配器，将返回结果写入 `review_result`（`check_type`=`ai`）
- AI 调用失败时记录日志，写入 `status=fail` 的结果，不影响编译结果展示
- 未配置 AI 服务商时跳过

```go
// services/ai_reviewer.go
func RunAIReview(task models.BuildTask) error {
    provider := beego.AppConfig.String("ai_provider") // "claude" or "openai"
    if provider == "" {
        return nil
    }
    var reviewer ai.AIReviewer
    switch provider {
    case "claude":
        reviewer = ai.NewClaudeReviewer(decryptKey(...))
    case "openai":
        reviewer = ai.NewOpenAIReviewer(decryptKey(...))
    }
    diff := getGitDiff(task)
    result, err := reviewer.Review(diff, promptTemplate)
    // 写入 ReviewResult
}
```

### 管理员配置入口

**文件：** `controllers/admin.go`

- 可配置：AI 服务商、API Key（输入后加密存储）、审查提示词
- API Key 界面展示脱敏（`****`）

---

## 验证清单

- [x] 支持 Claude 和 OpenAI 两种服务商
- [x] API Key 使用 AES-256-GCM 加密存储
- [x] 界面展示脱敏（`****`）
- [x] 编译完成后自动执行 AI 审查
- [x] 结果写入 `review_result` 表（check_type=ai）
- [x] AI 调用失败时记录日志，不影响编译结果
- [x] 未配置时跳过
- [x] 任务详情页显示 AI 审查报告折叠面板

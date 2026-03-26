---
story_id: 2-2
story_key: NOKEY-2-2
title: Git 提交规范检查
epic: epic-2
status: done
created: '2026-03-25'
---

# Story 2-2：Git 提交规范检查

## 目标

每次编译任务执行时自动检查 commit message 是否符合管理员配置的正则规范，不合规时在报告中标注，不阻断编译流程。

---

## 依赖

| 前置 Story | 依赖内容 |
|------------|----------|
| Story 1-1 | `ReviewResult` 模型（`check_type`=`git`）、系统配置表 |
| Story 1-5 | 编译任务执行流程 |

---

## 已实现代码上下文

### Git 规范检查服务

**文件：** `services/git_checker.go`

- `CheckCommitMessage(task BuildTask) error` — 用管理员配置的正则对 `task.CommitMsg` 做匹配
- 合规 → `ReviewResult{Status: pass}`；不合规 → `ReviewResult{Status: fail, Detail: 说明}`
- 结果写入 `review_result`（`check_type`=`git`）
- 未配置正则时跳过，不写入结果

```go
// services/git_checker.go
func CheckCommitMessage(task models.BuildTask) error {
    pattern := beego.AppConfig.String("commit_msg_pattern")
    if pattern == "" {
        return nil // 未配置，跳过
    }
    matched, _ := regexp.MatchString(pattern, task.CommitMsg)
    status := models.ReviewStatusPass
    detail := ""
    if !matched {
        status = models.ReviewStatusFail
        detail = fmt.Sprintf("commit message 不符合规范：%s", pattern)
    }
    // 写入 ReviewResult
}
```

### 管理员配置入口

**文件：** `controllers/admin.go`

- 管理员可在系统设置页配置 commit message 正则表达式
- 保存至 beego 配置或系统配置表

### 任务详情页展示

**文件：** `views/build/detail.tpl`

- Git 规范检查结果面板：合规 ✅ / 不合规 ❌ + 说明文字
- 与其他审查面板并列显示

---

## 验证清单

- [x] 每次任务执行时自动检查 commit message
- [x] 结果写入 `review_result` 表（check_type=git）
- [x] 不合规时标注，不阻断编译流程
- [x] 管理员可配置正则表达式
- [x] 未配置正则时跳过
- [x] 任务详情页显示 Git 规范检查结果

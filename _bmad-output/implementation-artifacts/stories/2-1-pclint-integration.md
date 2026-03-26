---
story_id: 2-1
story_key: NOKEY-2-1
title: PC-Lint 集成
epic: epic-2
status: done
created: '2026-03-25'
---

# Story 2-1：PC-Lint 集成

## 目标

编译成功后自动执行 PC-Lint 静态分析，将结果保存至数据库，并在任务详情页展示折叠报告面板。

---

## 依赖

| 前置 Story | 依赖内容 |
|------------|----------|
| Story 1-1 | `ReviewResult` 模型（`check_type`=`lint`）|
| Story 1-5 | 编译完成后执行点 |
| Story 1-7 | 任务详情页展示区域 |

---

## 已实现代码上下文

### Linter 服务

**文件：** `services/linter.go`

- `RunLint(task BuildTask) error` — 调用 PC-Lint 可执行文件对源码目录执行检查
- 解析 lint 输出，统计 warning/error 数量
- 将结果写入 `review_result`（`check_type`=`lint`，`status`=pass/warn/fail，`detail` 为 JSON）
- 仓库配置支持上传 `.lnt` 配置文件（≤1MB），存放至 `data/lint-configs/:repoName/`
- 未配置 lint 可执行路径时跳过，不影响编译结果

```go
// services/linter.go
func RunLint(task models.BuildTask) error {
    // 检查系统配置中 pc_lint_path 是否存在
    // 执行: pc_lint_path + lnt配置文件 + 源码路径
    // 解析输出，写入 ReviewResult
}
```

### 仓库配置（Lint 配置文件上传）

**文件：** `controllers/repo.go`

- `POST /repo/:repoName/lint-config` — 上传 `.lnt` 文件
- `DELETE /repo/:repoName/lint-config` — 删除，恢复为跳过 Lint
- 显示文件名和上传时间，提供系统默认模板下载链接

### 任务详情页展示

**文件：** `views/build/detail.tpl`

- PC-Lint 结果折叠面板：显示警告数/错误数
- 有问题时自动展开
- 可展开查看详细条目列表

---

## 验证清单

- [x] 编译成功后自动执行 PC-Lint
- [x] 结果写入 `review_result` 表（check_type=lint）
- [x] 任务详情页显示折叠面板（警告数/错误数）
- [x] 未配置 lint 路径时跳过，不报错
- [x] 仓库配置页支持上传 .lnt 文件
- [x]
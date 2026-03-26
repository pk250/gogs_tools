---
story_id: 3-4
story_key: NOKEY-3-4
title: PC-Lint 配置文件管理界面
epic: epic-3
status: done
created: '2026-03-25'
---

# Story 3-4：PC-Lint 配置文件管理界面

## 目标

开发者通过界面上传和管理 PC-Lint `.lnt` 配置文件，无需手动操作服务器文件系统，支持上传、查看、删除和下载默认模板。

---

## 依赖

| 前置 Story | 依赖内容 |
|------------|----------|
| Story 1-3 | 仓库配置页（UI 入口）|
| Story 2-1 | PC-Lint 服务（使用上传的 .lnt 文件）|

---

## 已实现代码上下文

### 上传/删除接口

**文件：** `controllers/repo.go`

```go
// POST /repo/:repoName/lint-config
// 上传 .lnt 文件，文件大小限制 ≤1MB
// 保存至 data/lint-configs/:repoName/config.lnt
func (this *RepoController) LintConfigUpload() {
    file, header, err := this.GetFile("lntFile")
    // 检查扩展名 .lnt、大小 ≤1MB
    // 保存至 data/lint-configs/:repoName/
    // 记录上传时间至 repo_config
}

// DELETE /repo/:repoName/lint-config
// 删除已上传的配置文件（恢复为跳过 Lint）
func (this *RepoController) LintConfigDelete() {
    os.Remove(lntPath)
    // 清除 repo_config 中的记录
}
```

### 仓库配置页 UI

**文件：** `views/repo/config.tpl`

- 显示当前已上传的 `.lnt` 文件名和上传时间
- 文件上传表单（接受 `.lnt` 文件，≤1MB）
- 「下载默认模板」链接
- 「删除配置文件」按钮（删除后恢复为跳过 Lint）

### 文件存储路径

```
data/
└── lint-configs/
    └── :repoName/
        └── config.lnt
```

---

## 验证清单

- [x] 仓库配置页支持上传 .lnt 文件
- [x] 文件大小限制 ≤1MB，超出时提示错误
- [x] 上传后显示文件名和上传时间
- [x] 提供系统默认模板下载链接
- [x] 可删除已上传的配置文件（恢复为跳过 Lint）
- [x] 删除后 PC-Lint 服务跳过该仓库的检查

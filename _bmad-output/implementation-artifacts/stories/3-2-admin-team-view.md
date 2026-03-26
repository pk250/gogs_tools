---
story_id: 3-2
story_key: NOKEY-3-2
title: 管理员团队视图
epic: epic-3
status: done
created: '2026-03-25'
---

# Story 3-2：管理员团队视图

## 目标

提供管理员专属页面，展示所有仓库的配置摘要和绑定用户，支持按成员筛选，可直接进入任意仓库配置编辑页。

---

## 依赖

| 前置 Story | 依赖内容 |
|------------|----------|
| Story 1-2 | Keil 版本列表 |
| Story 1-3 | 仓库编译配置模型 |

---

## 已实现代码上下文

### 团队视图控制器

**文件：** `controllers/admin.go`

- `GET /admin/team` — 列出所有仓库 + 配置摘要 + 推送用户列表
- 支持按 `member` 参数筛选，只显示该用户有推送记录的仓库
- 关联 Keil 版本名称显示

```go
// controllers/admin.go — TeamView()
func (this *AdminController) TeamView() {
    // 读取所有仓库名（从 gogs_d_b 推送记录）
    // 关联 repo_config + keil_version
    // 查询每个仓库的推送用户列表
    // 支持 filterMember 参数
    this.TplName = "admin/team.tpl"
}
```

每条记录结构：

```go
type repoEntry struct {
    RepoName  string
    Config    models.RepoConfig
    HasConfig bool
    KeilName  string
    Pushers   []string // 该仓库的推送用户列表
}
```

### 模板

**文件：** `views/admin/team.tpl`

- 表格展示：仓库名 / Keil版本 / 触发策略 / 推送用户 / 操作（进入配置）
- 成员筛选下拉框（从推送记录中取 distinct 用户名）
- 「进入配置」按钮直接跳转至该仓库的配置编辑页

---

## 验证清单

- [x] 管理员专属页面，非管理员重定向
- [x] 展示所有仓库列表 + 配置摘要 + 绑定用户
- [x] 按成员筛选正常工作
- [x] 可直接进入任意仓库的配置编辑页
- [x] 无配置的仓库也显示在列表中（HasConfig=false）

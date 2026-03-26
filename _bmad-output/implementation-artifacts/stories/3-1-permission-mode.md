---
story_id: 3-1
story_key: NOKEY-3-1
title: 角色权限模式切换
epic: epic-3
status: done
created: '2026-03-25'
---

# Story 3-1：角色权限模式切换

## 目标

管理员可在系统设置中切换权限模式（宽松/严格），控制仓库编译配置的修改权限，通过 middleware 强制执行，无需重启生效。

---

## 依赖

| 前置 Story | 依赖内容 |
|------------|----------|
| Story 1-3 | 仓库编译配置页（被权限保护的目标）|

---

## 已实现代码上下文

### 权限 Middleware

**文件：** `middleware/auth.go`

- 宽松模式：已登录用户均可修改仓库编译配置
- 严格模式：仅 `is_admin=true` 或 `role=project_lead` 的用户可修改
- 权限检查在服务端 middleware 执行，不依赖前端隐藏
- 读取系统配置中的 `permission_mode` 字段（`loose`/`strict`）

```go
// middleware/auth.go
func RequireConfigPermission(ctx *context.Context) {
    mode := getPermissionMode() // 从DB或配置读取
    if mode == "strict" {
        user := getSessionUser(ctx)
        if !user.IsAdmin && user.Role != "project_lead" {
            ctx.Redirect(302, "/dashboard")
        }
    }
}
```

### 管理员切换入口

**文件：** `controllers/admin.go`

- 系统设置页显示当前权限模式
- `POST /admin/settings` — 保存权限模式，即时写入配置存储
- 无需重启，下一次请求即生效

---

## 验证清单

- [x] 管理员可在系统设置中切换权限模式
- [x] 宽松模式：所有登录用户可配置仓库
- [x] 严格模式：仅 project_lead/admin 可修改仓库编译配置
- [x] 权限检查通过 middleware 实现，不依赖前端隐藏
- [x] 模式切换即时生效，无需重启

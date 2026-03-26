---
story_id: 1-1
story_key: NOKEY-1-1
title: 数据库模型与迁移
epic: epic-1
status: ready-for-dev
created: '2026-03-25'
---

# Story 1-1：数据库模型与迁移

## 目标

创建编译任务相关的数据库表和 Go 模型，为后续所有 Epic 1 Story 提供数据基础。

---

## 现有代码库上下文（必读）

### 现有模型注册方式

**文件：** `models/users.go`（第 24-34 行）

```go
func init() {
    orm.RegisterDataBase("default", "mysql", "root:123456@tcp(localhost:3306)/gogs?charset=utf8")
    orm.RegisterModel(new(Datainfos))
    orm.RegisterModel(new(Users))
    orm.RegisterModel(new(GogsDB))
    orm.RunSyncdb("default", false, true)
}
```

**关键规则：**
- 所有新模型必须在此 `init()` 函数中用 `orm.RegisterModel()` 注册
- `orm.RunSyncdb("default", false, true)` 已设置为 AutoMigrate 模式（第二参数 false=不强制重建，第三参数 true=打印 SQL）
- 数据库连接字符串在 `models/users.go` 中，**不要修改**

### 现有模型文件

| 文件 | 模型 | 说明 |
|------|------|------|
| `models/users.go` | `Users` | 用户表，含 IsAdmin/IsStaff 角色字段 |
| `models/datainfos.go` | `Datainfos` | 旧版编译记录表（保留，不修改）|
| `models/gogs.go` | `Commits`, `Repository`, `GogsDB` 等 | Webhook 数据结构 |

### 现有 Datainfos 表（保留不动）

`Datainfos` 表已存在并有数据，**不要修改或删除**。新模型是独立的新表。

---

## 验收标准

- [ ] 创建 `keil_versions` 表，对应 `models/keil_version.go`
- [ ] 创建 `repos` 表扩展，对应 `models/repo.go`
- [ ] 创建 `build_tasks` 表，对应 `models/build_task.go`
- [ ] 创建 `review_results` 表，对应 `models/review_result.go`
- [ ] 创建 `sys_config` 表，对应 `models/sys_config.go`
- [ ] 所有新模型在 `models/users.go` 的 `init()` 中注册
- [ ] `orm.RunSyncdb` 执行后无报错，新表成功创建
- [ ] 编写 `models/migrate.go`，提供手动迁移辅助函数（为现有 `datainfos` 表追加字段备用）

---

## 技术实现规范

### 文件结构

```
models/
├── users.go          ← 已有，修改 init() 追加注册
├── gogs.go           ← 已有，不修改
├── datainfos.go      ← 已有，不修改
├── keil_version.go   ← 新建
├── repo.go           ← 新建
├── build_task.go     ← 新建
├── review_result.go  ← 新建
├── sys_config.go     ← 新建
└── migrate.go        ← 新建（迁移辅助）
```

### 模型定义

#### `models/keil_version.go`

```go
package models

import "time"

type KeilVersion struct {
    Id          int64
    VersionName string    `orm:"size(64);unique"`
    Uv4Path     string    `orm:"size(512)"`
    CreatedAt   time.Time `orm:"auto_now_add;type(datetime)"`
    UpdatedAt   time.Time `orm:"auto_now;type(datetime)"`
}
```

#### `models/repo.go`

```go
package models

import "time"

// RepoConfig 仓库编译配置表
// 通过 RepoName 与 GogsDB.Repository_Name 关联（非外键，避免耦合）
type RepoConfig struct {
    Id             int64
    RepoName       string       `orm:"size(128);unique"`  // 仓库名，唯一
    KeilVersionId  int64        `orm:"default(0)"`        // 关联 KeilVersion.Id
    TriggerMode    string       `orm:"size(16);default(manual)"` // auto | manual
    ArtifactName   string       `orm:"size(128);null"`    // 产物文件名，默认仓库名
    NotifyEmails   string       `orm:"type(text);null"`   // 逗号分隔的邮件列表
    WebhookEnabled bool         `orm:"default(false)"`
    WebhookUrl     string       `orm:"size(512);null"`
    LintConfigPath string       `orm:"size(512);null"`    // .lnt 文件服务器路径
    CreatedAt      time.Time    `orm:"auto_now_add;type(datetime)"`
    UpdatedAt      time.Time    `orm:"auto_now;type(datetime)"`
}
```

#### `models/build_task.go`

```go
package models

import "time"

// BuildTaskStatus 任务状态枚举
const (
    TaskStatusPending = "pending"
    TaskStatusRunning = "running"
    TaskStatusSuccess = "success"
    TaskStatusFailed  = "failed"
)

type BuildTask struct {
    Id           int64
    RepoName     string    `orm:"size(128)"`          // 仓库名
    CommitHash   string    `orm:"size(64)"`           // commit hash
    CommitMsg    string    `orm:"type(text);null"`    // commit message
    Author       string    `orm:"size(128);null"`     // 提交人
    Status       string    `orm:"size(16);default(pending)"` // pending|running|success|failed
    LogPath      string    `orm:"size(512);null"`     // 日志文件路径
    ErrorSummary string    `orm:"type(text);null"`    // 错误摘要（关键错误行）
    StartedAt    time.Time `orm:"null;type(datetime)"`
    FinishedAt   time.Time `orm:"null;type(datetime)"`
    CreatedAt    time.Time `orm:"auto_now_add;type(datetime)"`
}
```

#### `models/review_result.go`

```go
package models

import "time"

// ReviewType 审查类型
const (
    ReviewTypeLint    = "lint"
    ReviewTypeAI      = "ai"
    ReviewTypeGitMsg  = "git_msg"
)

type ReviewResult struct {
    Id         int64
    TaskId     int64     `orm:"index"`              // 关联 BuildTask.Id
    ReviewType string    `orm:"size(32)"`            // lint | ai | git_msg
    Status     string    `orm:"size(16)"`            // pass | warn | fail | skip
    Summary    string    `orm:"type(text);null"`     // 结果摘要
    Detail     string    `orm:"type(longtext);null"` // 详细内容（JSON 或纯文本）
    CreatedAt  time.Time `orm:"auto_now_add;type(datetime)"`
}
```

#### `models/sys_config.go`

```go
package models

import "time"

// SysConfig 系统配置 KV 表
// 敏感值（AI Key、SMTP密码）存储时须加密，通过 ConfigService 读写
type SysConfig struct {
    Id        int64
    ConfigKey string    `orm:"size(128);unique"`
    ConfigVal string    `orm:"type(text);null"`   // 加密值或明文值
    IsSecret  bool      `orm:"default(false)"`    // true=加密存储
    UpdatedAt time.Time `orm:"auto_now;type(datetime)"`
}

// 预定义 ConfigKey 常量，避免硬编码字符串
const (
    ConfigKeyAIProvider   = "ai_provider"    // claude | openai
    ConfigKeyAIApiKey     = "ai_api_key"     // 加密存储
    ConfigKeyAIPrompt     = "ai_prompt"      // 审查提示词
    ConfigKeySMTPHost     = "smtp_host"
    ConfigKeySMTPPort     = "smtp_port"
    ConfigKeySMTPUser     = "smtp_user"
    ConfigKeySMTPPass     = "smtp_pass"      // 加密存储
    ConfigKeyGitMsgRegex  = "git_msg_regex"  // commit message 规范正则
    ConfigKeyPermMode     = "perm_mode"      // loose | strict
)
```

### 修改 `models/users.go` 的 init()

在现有 `orm.RegisterModel(new(GogsDB))` 之后，追加以下注册（**不要修改其他任何内容**）：

```go
// 新增：编译流水线相关模型
orm.RegisterModel(new(KeilVersion))
orm.RegisterModel(new(RepoConfig))
orm.RegisterModel(new(BuildTask))
orm.RegisterModel(new(ReviewResult))
orm.RegisterModel(new(SysConfig))
```

### `models/migrate.go`（迁移辅助）

```go
package models

import (
    "github.com/astaxie/beego/orm"
    "github.com/astaxie/beego/logs"
)

// RunMigrations 执行手动迁移（服务启动时调用）
// 目前为空，预留给未来需要手动 ALTER TABLE 的场景
func RunMigrations() {
    o := orm.NewOrm()
    _ = o
    logs.Info("[migrate] 数据库迁移检查完成")
}
```

---

## 关键约束（开发 Agent 必须遵守）

| 约束 | 说明 |
|------|------|
| 不修改现有表 | `datainfos`、`users`、`gogs_db` 表结构不变 |
| 不修改连接字符串 | 数据库连接在 `models/users.go` 第25行，不要改 |
| 保持 init() 结构 | 只追加 RegisterModel，不改现有调用顺序 |
| 状态值用常量 | 使用 `TaskStatusPending` 等常量，不硬编码字符串 |
| ConfigKey 用常量 | 使用 `ConfigKeyAIApiKey` 等常量 |
| 不实现业务逻辑 | 本 Story 只建模型和表，不实现 Service/Controller |

---

## 测试验证步骤

1. 运行 `go build ./...` — 无编译错误
2. 启动服务 `go run main.go` — 观察启动日志，确认 Beego ORM 输出以下建表 SQL：
   - `CREATE TABLE IF NOT EXISTS keil_version`
   - `CREATE TABLE IF NOT EXISTS repo_config`
   - `CREATE TABLE IF NOT EXISTS build_task`
   - `CREATE TABLE IF NOT EXISTS review_result`
   - `CREATE TABLE IF NOT EXISTS sys_config`
3. 连接 MySQL 确认 5 张新表已创建，字段符合规范
4. 确认现有表（`datainfos`、`users`、`gogs_db`）结构未变化

---

## 完成后下一步

Story 1-1 完成后，可并行开始：
- **Story 1-2**：Keil 版本管理（依赖 `KeilVersion` 模型）
- **Story 1-3**：仓库编译配置（依赖 `RepoConfig` 模型）

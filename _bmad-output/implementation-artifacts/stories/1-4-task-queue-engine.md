---
story_id: 1-4
story_key: NOKEY-1-4
title: 任务队列引擎
epic: epic-1
status: ready-for-dev
created: '2026-03-25'
---

# Story 1-4：任务队列引擎

## 目标

实现持久化的并发任务队列，支持最多 5 个并发 Worker、队列上限 50、服务重启后自动恢复未完成任务，并提供手动触发接口。

---

## 依赖

- **Story 1-1 必须先完成**：`models/build_task.go` 中的 `BuildTask` 模型和状态常量必须已存在
- **Story 1-3 必须先完成**：`models/repo.go` 中的 `RepoConfig`（`TriggerMode` 字段）必须已存在，用于判断自动/手动模式

---

## 现有代码库上下文（必读）

### main.go 当前结构

**文件：** `main.go`

```go
package main

import (
    _ "gogs_tools/routers"
    "github.com/astaxie/beego"
)

func main() {
    beego.Run()
}
```

**关键规则：** Dispatcher 必须在 `beego.Run()` 之前启动，否则 HTTP 请求到达时队列尚未就绪。修改 `main.go` 追加 `services.StartDispatcher()`。

### controllers/gogs.go 现有 Webhook 处理

**文件：** `controllers/gogs.go`（第 15-58 行）

现有逻辑：接收 Webhook → 插入 `GogsDB` → 插入 `Datainfos`（旧版）。
**本 Story 要求：** 在插入 `Datainfos` 之后，判断该仓库的 `TriggerMode`，若为 `auto` 则调用 `QueueService.Enqueue()`。
**不要删除**现有的 `Datainfos` 插入逻辑，保持向后兼容。

### BuildTask 模型（Story 1-1 创建）

**文件：** `models/build_task.go`

```go
const (
    TaskStatusPending = "pending"
    TaskStatusRunning = "running"
    TaskStatusSuccess = "success"
    TaskStatusFailed  = "failed"
)

type BuildTask struct {
    Id           int64
    RepoName     string    `orm:"size(128)"`
    CommitHash   string    `orm:"size(64)"`
    CommitMsg    string    `orm:"type(text);null"`
    Author       string    `orm:"size(128);null"`
    Status       string    `orm:"size(16);default(pending)"`
    LogPath      string    `orm:"size(512);null"`
    ExitCode     int       `orm:"default(-1)"`
    StartedAt    time.Time `orm:"null;type(datetime)"`
    FinishedAt   time.Time `orm:"null;type(datetime)"`
    CreatedAt    time.Time `orm:"auto_now_add;type(datetime)"`
    UpdatedAt    time.Time `orm:"auto_now;type(datetime)"`
}
```

### 认证机制（继承 BaseController）

**文件：** `controllers/base.go`

所有需要登录的接口都必须继承 `BaseController`，`Prepare()` 自动检查 Session。手动触发接口 `POST /api/build/trigger` 需要登录认证，直接继承即可。

### 路由注册方式

**文件：** `routers/router.go`

现有路由示例：
```go
beego.Router("/api", &controllers.ApiController{})
beego.Router("/gogs", &controllers.GogsControllers{})
```

新路由需追加在 `init()` 末尾。

---

## 验收标准

- [ ] Webhook 自动触发后任务入队，Webhook HTTP 响应时间 ≤500ms（入队本身是同步快速操作）
- [ ] 最多 5 个并发 Worker Goroutine，第 6 个任务排队等待
- [ ] 队列上限 50，超出时返回 HTTP 429 + `{"code":429,"message":"队列已满，请稍后重试"}`
- [ ] 服务启动时扫描 `status=running` 的任务，重置为 `pending` 重新入队
- [ ] 任务状态变更仅通过 `QueueService.UpdateStatus()` 单一入口
- [ ] 手动触发接口：`POST /api/build/trigger`，需登录认证
- [ ] Dispatcher goroutine 在 `beego.Run()` 之前启动

---

## 技术实现规范

### 文件结构

```
services/
├── queue.go          ← 新建：QueueService（Enqueue、UpdateStatus、Recover）
└── worker.go         ← 新建：Dispatcher 和 Worker goroutine

controllers/
└── build.go          ← 新建：BuildController（手动触发接口）

修改文件：
├── main.go           ← 追加 Dispatcher 启动
├── routers/router.go ← 追加路由
└── controllers/gogs.go ← 追加自动入队逻辑
```

### `services/queue.go` 实现规范

```go
package services

import (
    "fmt"
    "time"

    "gogs_tools/models"

    "github.com/astaxie/beego/orm"
)

const (
    MaxWorkers  = 5
    MaxQueueCap = 50
)

// TaskCh 是全局任务 channel，Dispatcher 从中消费
// buffered channel 容量 = MaxQueueCap
var TaskCh = make(chan int64, MaxQueueCap)

// Enqueue 创建任务并入队，返回 task ID
// 若 channel 已满，返回 error（调用方返回 HTTP 429）
func Enqueue(repoName, commitHash, commitMsg, author string) (int64, error) {
    // 先检查 channel 是否已满（非阻塞）
    if len(TaskCh) >= MaxQueueCap {
        return 0, fmt.Errorf("queue full")
    }

    // INSERT build_tasks
    o := orm.NewOrm()
    task := &models.BuildTask{
        RepoName:   repoName,
        CommitHash: commitHash,
        CommitMsg:  commitMsg,
        Author:     author,
        Status:     models.TaskStatusPending,
    }
    id, err := o.Insert(task)
    if err != nil {
        return 0, fmt.Errorf("insert build_task: %w", err)
    }

    // 非阻塞写入 channel
    select {
    case TaskCh <- id:
    default:
        // channel 刚好在检查后满了，任务已在 DB，Recover 会处理
    }

    return id, nil
}

// UpdateStatus 是任务状态变更的唯一入口
// 禁止在其他地方直接 UPDATE build_tasks.status
func UpdateStatus(taskId int64, status string) error {
    o := orm.NewOrm()
    task := &models.BuildTask{Id: taskId}
    if err := o.Read(task); err != nil {
        return fmt.Errorf("read task %d: %w", taskId, err)
    }
    task.Status = status
    if status == models.TaskStatusRunning {
        task.StartedAt = time.Now()
    } else if status == models.TaskStatusSuccess || status == models.TaskStatusFailed {
        task.FinishedAt = time.Now()
    }
    _, err := o.Update(task)
    return err
}

// Recover 在服务启动时调用，将 running→pending，并重新入队 pending 任务
// 调用时机：main.go 中 beego.Run() 之前
func Recover() {
    o := orm.NewOrm()

    // 1. running → pending（上次崩溃未完成的任务）
    o.Raw("UPDATE build_task SET status=? WHERE status=?",
        models.TaskStatusPending, models.TaskStatusRunning).Exec()

    // 2. 将所有 pending 任务重新入队（按 ID 排序）
    var tasks []*models.BuildTask
    o.QueryTable("build_task").
        Filter("Status", models.TaskStatusPending).
        OrderBy("Id").
        Limit(MaxQueueCap).
        All(&tasks)

    for _, t := range tasks {
        select {
        case TaskCh <- t.Id:
        default:
            // channel 已满，剩余任务等待 Dispatcher 轮询
        }
    }
}
```

**ORM 表名注意：** Beego ORM 将 `BuildTask` 自动映射为 `build_task`，Raw SQL 中必须使用 `build_task`。

### `services/worker.go` 实现规范

```go
package services

import (
    "context"
    "sync"

    "github.com/astaxie/beego/logs"
)

// StartDispatcher 启动 Dispatcher，在 main.go 中 beego.Run() 之前调用
// ctx 用于优雅关闭（当前 MVP 阶段传 context.Background()）
func StartDispatcher(ctx context.Context) {
    var wg sync.WaitGroup
    sem := make(chan struct{}, MaxWorkers) // 信号量控制并发数

    go func() {
        for {
            select {
            case <-ctx.Done():
                wg.Wait() // 等待所有 Worker 完成
                return
            case taskId := <-TaskCh:
                sem <- struct{}{} // 占用一个 Worker 槽位
                wg.Add(1)
                go func(id int64) {
                    defer wg.Done()
                    defer func() { <-sem }() // 释放槽位
                    runWorker(id)
                }(taskId)
            }
        }
    }()

    logs.Info("[Queue] Dispatcher started, maxWorkers=%d, queueCap=%d", MaxWorkers, MaxQueueCap)
}

// runWorker 执行单个任务
// MVP 阶段：仅更新状态为 running，实际编译由 Story 1-5 实现
// 本 Story 只需要骨架，Story 1-5 会填充编译逻辑
func runWorker(taskId int64) {
    logs.Info("[Worker] Start task %d", taskId)

    if err := UpdateStatus(taskId, "running"); err != nil {
        logs.Error("[Worker] UpdateStatus running, task=%d, err=%v", taskId, err)
        return
    }

    // TODO: Story 1-5 在此处调用 compiler.Run(taskId)
    // 当前 MVP 骨架：直接标记 success
    // 实际实现时删除下面这行，替换为 compiler.Run()
    if err := UpdateStatus(taskId, "success"); err != nil {
        logs.Error("[Worker] UpdateStatus success, task=%d, err=%v", taskId, err)
    }

    logs.Info("[Worker] Done task %d", taskId)
}
```

**关键：** `runWorker` 当前是骨架实现（直接 success），Story 1-5 实现时会替换

### `controllers/build.go` 实现规范

```go
package controllers

import (
    "gogs_tools/models"
    "gogs_tools/services"

    "github.com/astaxie/beego/orm"
)

type BuildController struct {
    BaseController  // 继承登录鉴权
}

// Trigger POST /api/build/trigger
// 请求体 JSON: {"repo_name":"myrepo","commit_hash":"abc123","commit_msg":"fix bug","author":"user"}
// 响应: {"code":0,"data":{"task_id":42}} 或 {"code":429,"message":"队列已满"}
func (this *BuildController) Trigger() {
    type triggerReq struct {
        RepoName   string `json:"repo_name"`
        CommitHash string `json:"commit_hash"`
        CommitMsg  string `json:"commit_msg"`
        Author     string `json:"author"`
    }

    var req triggerReq
    if err := this.Ctx.Input.BindJSON(&req); err != nil || req.RepoName == "" {
        this.Ctx.ResponseWriter.WriteHeader(400)
        this.Data["json"] = map[string]interface{}{"code": 400, "message": "参数错误"}
        this.ServeJSON()
        return
    }

    taskId, err := services.Enqueue(req.RepoName, req.CommitHash, req.CommitMsg, req.Author)
    if err != nil {
        if err.Error() == "queue full" {
            this.Ctx.ResponseWriter.WriteHeader(429)
            this.Data["json"] = map[string]interface{}{"code": 429, "message": "队列已满，请稍后重试"}
        } else {
            this.Ctx.ResponseWriter.WriteHeader(500)
            this.Data["json"] = map[string]interface{}{"code": 500, "message": "内部错误"}
        }
        this.ServeJSON()
        return
    }

    this.Data["json"] = map[string]interface{}{"code": 0, "data": map[string]interface{}{"task_id": taskId}}
    this.ServeJSON()
}
```

**注意：** Beego `BindJSON` 需要请求 `Content-Type: application/json`，且请求体已由框架读取（`EnableHTTPBody = true`）。若 BindJSON 不可用，改用 `json.Unmarshal(this.Ctx.Input.RequestBody, &req)`（参考 `controllers/gogs.go` 第 18-20 行的写法）。

### `main.go` 修改规范

```go
package main

import (
    "context"
    _ "gogs_tools/routers"
    "gogs_tools/services"

    "github.com/astaxie/beego"
)

func main() {
    services.Recover()                              // 1. 恢复上次未完成任务
    services.StartDispatcher(context.Background())  // 2. 启动 Dispatcher
    beego.Run()                                     // 3. 启动 HTTP 服务
}
```

### `routers/router.go` 修改规范

在 `init()` 末尾追加：

```go
beego.Router("/api/build/trigger", &controllers.BuildController{}, "post:Trigger")
```

### `controllers/gogs.go` 修改规范

在现有 `for _, commit := range gogs.Commits` 循环内，`o.Insert(&datainfo)` 成功后追加自动入队逻辑：

```go
// 判断触发模式，自动模式则入队
repoConfig := models.RepoConfig{RepoName: gogs.Repository.Name}
err2 := o.Read(&repoConfig, "RepoName")
if err2 == nil && repoConfig.TriggerMode == "auto" {
    services.Enqueue(
        gogs.Repository.Name,
        commit.Id,
        commit.Message,
        commit.Author.Name,
    )
}
```

**注意：** `Enqueue` 错误在此处可忽略（队列满时不影响 Webhook 响应），但建议 `logs.Warn` 记录。Import 需追加 `"gogs_tools/services"`。

---

## 关键陷阱与注意事项

| 陷阱 | 说明 | 解决方案 |
|------|------|----------|
| Dispatcher 未启动 | `beego.Run()` 后才调用 `StartDispatcher` → 服务启动时队列未就绪 | 严格按 Recover → StartDispatcher → beego.Run 顺序 |
| channel 满判断竞态 | `len(TaskCh) >= MaxQueueCap` 检查后 channel 可能刚好写满 | select default 兜底：任务已在 DB，Recover 会处理 |
| ORM 表名 | `BuildTask` → `build_task`（单下划线）；`GogsDB` → `gogs_d_b`（双下划线）| Raw SQL 必须用 `build_task` |
| BindJSON 不可用 | Beego 版本差异，BindJSON 可能不存在 | 改用 `json.Unmarshal(this.Ctx.Input.RequestBody, &req)` |
| 并发 Worker 数量 | 信号量 `sem` 大小即最大并发数，channel 容量即等待队列大小 | `sem := make(chan struct{}, MaxWorkers)` 不要搞反 |
| Recover 时队列容量 | `Limit(MaxQueueCap)` 避免历史 pending 任务超过 channel 容量 | 已在规范中处理 |

---

## 测试验证方法

```bash
# 1. 手动触发（需先登录获取 cookie）
curl -X POST http://localhost:8080/api/build/trigger \
  -H 'Content-Type: application/json' \
  -b 'BEEGOSESSIONID=xxx' \
  -d '{"repo_name":"testrepo","commit_hash":"abc123","commit_msg":"test","author":"dev"}'

# 期望响应: {"code":0,"data":{"task_id":1}}

# 2. 验证数据库
# SELECT * FROM build_task ORDER BY id DESC LIMIT 5;

# 3. 验证队列满（发送 51 个请求）
# 第 51 个请求应返回 HTTP 429
```

---

## 交付清单

- [ ] `services/queue.go` — Enqueue、UpdateStatus、Recover 三个函数
- [ ] `services/worker.go` — StartDispatcher、runWorker（骨架）
- [ ] `controllers/build.go` — BuildController.Trigger
- [ ] `main.go` — 追加 Recover() 和 StartDispatcher()
- [ ] `routers/router.go` — 追加 /api/build/trigger 路由
- [ ] `controllers/gogs.go` — 追加自动入队逻辑
- [ ] 启动服务无报错，日志显示 `[Queue] Dispatcher started`
- [ ] 手动触发接口返回 task_id，DB 中出现对应记录
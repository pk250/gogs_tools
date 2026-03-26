---
story_id: 1-5
story_key: NOKEY-1-5
title: Keil 编译进程执行
epic: epic-1
status: ready-for-dev
created: '2026-03-25'
---

# Story 1-5：Keil 编译进程执行

## 目标

实现 `services/compiler.go`，调用 UV4.exe 执行编译，捕获输出，写入日志文件，复制产物，并替换 Story 1-4 中的 `runWorker` 骨架实现。

---

## 依赖关系

| 依赖 | Story | 原因 |
|------|-------|------|
| `BuildTask` 模型 | Story 1-1 | 读取任务信息、更新状态 |
| `KeilVersion` 模型 | Story 1-1 | 读取 UV4.exe 路径 |
| `RepoConfig` 模型 | Story 1-1 | 读取仓库编译配置（uvprojx路径、产物名）|
| `services.UpdateStatus()` | Story 1-4 | 唯一状态变更入口 |
| `services/worker.go` | Story 1-4 | 替换 runWorker 骨架 |

**注意：** Story 1-6 定义 WebSocket Hub，本 Story 通过接口预留广播钩子，不强依赖 Hub 实现。

---

## 现有代码库上下文（必读）

### 相关模型字段

**`models/keil_version.go`（Story 1-1 创建）**
```go
type KeilVersion struct {
    Id          int64
    VersionName string    `orm:"size(64);unique"`
    Uv4Path     string    `orm:"size(512)"`     // UV4.exe 完整路径，如 C:\Keil_v5\UV4\UV4.exe
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

**`models/repo.go`（Story 1-1 创建）**
```go
type RepoConfig struct {
    Id             int64
    RepoName       string    // 仓库名
    KeilVersionId  int64     // 关联 KeilVersion.Id
    ArtifactName   string    // 产物文件名（不含扩展名），默认仓库名
    // ...其他字段
}
```

**`models/build_task.go`（Story 1-1 创建）**
```go
type BuildTask struct {
    Id           int64
    RepoName     string
    CommitHash   string
    Status       string    // pending|running|success|failed
    LogPath      string    // 日志文件路径
    ErrorSummary string    // 错误摘要
    StartedAt    time.Time
    FinishedAt   time.Time
}
```

### Story 1-4 的 runWorker 骨架（需替换）

**文件：** `services/worker.go`

```go
// 当前骨架 — Story 1-5 实现时替换此函数体
func runWorker(task models.BuildTask) {
    UpdateStatus(task.Id, models.TaskStatusRunning)
    // TODO: 调用 compiler.Run(task)
    UpdateStatus(task.Id, models.TaskStatusSuccess)
}
```

**替换后应为：**
```go
func runWorker(task models.BuildTask) {
    UpdateStatus(task.Id, models.TaskStatusRunning)
    err := compiler.Run(task)
    if err != nil {
        UpdateStatus(task.Id, models.TaskStatusFailed)
    } else {
        UpdateStatus(task.Id, models.TaskStatusSuccess)
    }
}
```

### Story 1-6 预留的广播接口

Story 1-6 会实现 `services/ws_hub.go`，提供全局 Hub。本 Story 在 `compiler.go` 中通过函数变量预留钩子，Story 1-6 实现后再绑定：

```go
// services/compiler.go 中预留
// BroadcastLog 是可注入的日志广播函数，默认 noop，Story 1-6 绑定实现
var BroadcastLog func(taskId int64, line string) = func(int64, string) {}
```

---

## 验收标准

- [ ] `services/compiler.go` 实现 `Run(task BuildTask) error`
- [ ] 从 DB 读取 `KeilVersion.Uv4Path` 和仓库 uvprojx 文件路径
- [ ] 使用 `os/exec` + `context.WithTimeout`（90分钟）调用 UV4.exe
- [ ] stdout + stderr 同时写入 `/data/logs/{taskId}.log` 和 BroadcastLog 钩子
- [ ] 退出码 0 → success，非 0 → failed，`ErrorSummary` 存储最后 20 行日志
- [ ] 编译产物（.axf/.hex/.bin/.map）复制至 `/data/artifacts/{taskId}/`
- [ ] 超时强制 kill 进程并标记 failed，日志追加「编译超时，已终止」
- [ ] `services/worker.go` 中 `runWorker` 骨架替换为调用 `compiler.Run()`

---

## 技术实现规范

### 文件结构

```
services/
├── queue.go      ← Story 1-4 已有
├── worker.go     ← Story 1-4 已有，修改 runWorker
└── compiler.go   ← 新建（本 Story 核心）

data/             ← 运行时目录，程序启动时创建
├── logs/
│   └── {taskId}.log
├── artifacts/
│   └── {taskId}/
│       ├── firmware.axf
│       ├── firmware.hex
│       └── ...
└── repos/        ← 仓库本地克隆目录
    └── {repoName}/
        └── *.uvprojx
```

### uvprojx 文件查找策略

仓库目录为 `/data/repos/{repoName}/`（需在 `sys_config` 中配置 `repos_base_path`，默认 `./data/repos`）。uvprojx 文件通过递归搜索找到第一个 `.uvprojx` 文件：

```go
func findUvprojx(repoDir string) (string, error) {
    var found string
    err := filepath.Walk(repoDir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if !info.IsDir() && strings.HasSuffix(info.Name(), ".uvprojx") {
            found = path
            return filepath.SkipAll  // Go 1.20 支持
        }
        return nil
    })
    if err != nil {
        return "", fmt.Errorf("搜索 uvprojx 失败: %w", err)
    }
    if found == "" {
        return "", fmt.Errorf("仓库 %s 中未找到 .uvprojx 文件", repoDir)
    }
    return found, nil
}
```

**注意：** `filepath.SkipAll` 是 Go 1.20 引入的，项目已使用 Go 1.20，可安全使用。若不确定，可用自定义 sentinel error 替代。

### `services/compiler.go` 完整实现规范

```go
package services

import (
    "bufio"
    "context"
    "fmt"
    "io"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "time"

    "gogs_tools/models"

    "github.com/astaxie/beego/logs"
    "github.com/astaxie/beego/orm"
)

const (
    CompileTimeout  = 90 * time.Minute
    LogsBaseDir     = "./data/logs"
    ArtifactsDir    = "./data/artifacts"
    ReposBaseDir    = "./data/repos"
    ArtifactExts    = ".axf .hex .bin .map"
)

// BroadcastLog 可注入的 WebSocket 广播钩子，Story 1-6 绑定实现
// 默认 noop，编译日志通过此接口实时推送至前端
var BroadcastLog func(taskId int64, line string) = func(int64, string) {}

// Run 执行 Keil 编译流程，返回 nil 表示编译成功（退出码 0）
func Run(task models.BuildTask) error {
    o := orm.NewOrm()

    // 1. 读取仓库配置
    repoConfig := models.RepoConfig{RepoName: task.RepoName}
    if err := o.Read(&repoConfig, "RepoName"); err != nil {
        return fmt.Errorf("读取仓库配置失败: %w", err)
    }

    // 2. 读取 Keil 版本
    keilVersion := models.KeilVersion{Id: repoConfig.KeilVersionId}
    if err := o.Read(&keilVersion); err != nil {
        return fmt.Errorf("读取 Keil 版本失败 (id=%d): %w", repoConfig.KeilVersionId, err)
    }

    // 3. 查找 uvprojx 文件
    repoDir := filepath.Join(ReposBaseDir, task.RepoName)
    uvprojxPath, err := findUvprojx(repoDir)
    if err != nil {
        return err
    }

    // 4. 准备日志文件
    if err := os.MkdirAll(LogsBaseDir, 0755); err != nil {
        return fmt.Errorf("创建日志目录失败: %w", err)
    }
    logPath := filepath.Join(LogsBaseDir, fmt.Sprintf("%d.log", task.Id))
    logFile, err := os.Create(logPath)
    if err != nil {
        return fmt.Errorf("创建日志文件失败: %w", err)
    }
    defer logFile.Close()

    // 5. 更新日志路径到 DB
    o.QueryTable("build_task").Filter("Id", task.Id).Update(orm.Params{"LogPath": logPath})

    // 6. 执行编译（含超时）
    ctx, cancel := context.WithTimeout(context.Background(), CompileTimeout)
    defer cancel()

    // UV4.exe -rebuild {uvprojx} -j0 (j0=并行编译线程数不限)
    cmd := exec.CommandContext(ctx, keilVersion.Uv4Path, "-rebuild", uvprojxPath, "-j0")

    // stdout + stderr 合并到同一个 pipe
    stdoutPipe, err := cmd.StdoutPipe()
    if err != nil {
        return fmt.Errorf("获取 stdout pipe 失败: %w", err)
    }
    cmd.Stderr = cmd.Stdout  // stderr 合并到 stdout

    if err := cmd.Start(); err != nil {
        return fmt.Errorf("启动 UV4.exe 失败: %w", err)
    }

    // 7. 流式读取日志，同时写文件和广播
    var last20Lines []string
    scanner := bufio.NewScanner(stdoutPipe)
    for scanner.Scan() {
        line := scanner.Text()
        fmt.Fprintln(logFile, line)
        BroadcastLog(task.Id, line)  // Story 1-6 绑定后生效
        last20Lines = append(last20Lines, line)
        if len(last20Lines) > 20 {
            last20Lines = last20Lines[1:]
        }
    }

    cmdErr := cmd.Wait()

    // 8. 处理超时
    if ctx.Err() == context.DeadlineExceeded {
        timeoutMsg := "[gogs_tools] 编译超时（90分钟），已强制终止"
        fmt.Fprintln(logFile, timeoutMsg)
        BroadcastLog(task.Id, timeoutMsg)
        o.QueryTable("build_task").Filter("Id", task.Id).Update(orm.Params{
            "error_summary": timeoutMsg,
            "finished_at":   time.Now(),
        })
        return fmt.Errorf("编译超时")
    }

    // 9. 更新 ErrorSummary（最后 20 行）
    summary := strings.Join(last20Lines, "\n")
    finishedAt := time.Now()
    o.QueryTable("build_task").Filter("Id", task.Id).Update(orm.Params{
        "error_summary": summary,
        "finished_at":   finishedAt,
    })

    if cmdErr != nil {
        logs.Error("[compiler] 任务 %d 编译失败，退出码非0: %v", task.Id, cmdErr)
        return fmt.Errorf("编译失败: %w", cmdErr)
    }

    // 10. 复制产物
    if err := copyArtifacts(task, repoDir, repoConfig.ArtifactName); err != nil {
        logs.Warn("[compiler] 任务 %d 复制产物失败（不影响状态）: %v", task.Id, err)
    }

    return nil
}

// copyArtifacts 复制编译产物到 /data/artifacts/{taskId}/
func copyArtifacts(task models.BuildTask, repoDir, artifactName string) error {
    destDir := filepath.Join(ArtifactsDir, fmt.Sprintf("%d", task.Id))
    if err := os.MkdirAll(destDir, 0755); err != nil {
        return fmt.Errorf("创建产物目录失败: %w", err)
    }
    exts := strings.Fields(ArtifactExts)
    var copied int
    filepath.Walk(repoDir, func(path string, info os.FileInfo, err error) error {
        if err != nil || info.IsDir() {
            return nil
        }
        ext := strings.ToLower(filepath.Ext(path))
        for _, e := range exts {
            if ext == e {
                if copyErr := copyFile(path, filepath.Join(destDir, info.Name())); copyErr == nil {
                    copied++
                }
            }
        }
        return nil
    })
    if copied == 0 {
        return fmt.Errorf("未找到任何产物文件（.axf/.hex/.bin/.map）")
    }
    return nil
}

func copyFile(src, dst string) error {
    in, err := os.Open(src)
    if err != nil {
        return err
    }
    defer in.Close()
    out, err := os.Create(dst)
    if err != nil {
        return err
    }
    defer out.Close()
    _, err = io.Copy(out, in)
    return err
}

func findUvprojx(repoDir string) (string, error) {
    var found string
    filepath.Walk(repoDir, func(path string, info os.FileInfo, err error) error {
        if err != nil || info.IsDir() || found != "" {
            return nil
        }
        if strings.HasSuffix(strings.ToLower(info.Name()), ".uvprojx") {
            found = path
        }
        return nil
    })
    if found == "" {
        return "", fmt.Errorf("仓库 %s 中未找到 .uvprojx 文件", repoDir)
    }
    return found, nil
}
```

**io.Pipe 合并 stdout+stderr（推荐实现）：**

```go
pr, pw := io.Pipe()
cmd.Stdout = pw
cmd.Stderr = pw
cmd.Start()
go func() {
    cmd.Wait()
    pw.Close()
}()
scanner := bufio.NewScanner(pr)
```

---

## `services/worker.go` 修改规范

将 Story 1-4 的骨架替换为：

```go
func runWorker(task models.BuildTask) {
    UpdateStatus(task.Id, models.TaskStatusRunning)
    err := Run(task)  // services.Run，同 package
    if err != nil {
        logs.Error("[worker] 任务 %d 失败: %v", task.Id, err)
        UpdateStatus(task.Id, models.TaskStatusFailed)
    } else {
        UpdateStatus(task.Id, models.TaskStatusSuccess)
    }
}
```

---

## 关键陷阱

| 陷阱 | 说明 |
|------|------|
| `cmd.Stderr = cmd.Stdout` 在 StdoutPipe 后无效 | 用 `io.Pipe` + `cmd.Stdout = pw; cmd.Stderr = pw` 替代 |
| `orm.Params` key 是列名（蛇形）| `"finished_at"` 而非 `"FinishedAt"` |
| `filepath.SkipAll` | Go 1.20 才有，若编译报错改用自定义 sentinel error |
| `data/` 目录须提前创建 | 在 `compiler.go` 的 `init()` 或 `main.go` 中 `os.MkdirAll` |
| 产物复制失败不阻断状态 | `copyArtifacts` 错误仅 Warn 日志，不返回 error |

---

## 验收检查点

| 检查点 | 验证方式 |
|--------|----------|
| 日志文件生成 | `data/logs/{taskId}.log` 存在且非空 |
| 超时处理 | 传入不存在的 uvprojx，任务标记 failed（非 running 卡死）|
| 产物复制 | 编译成功后 `data/artifacts/{taskId}/` 含 .hex/.axf |
| worker 集成 | 手动触发后 build_tasks 状态 pending→running→success/failed |
| BroadcastLog 可注入 | Story 1-6 执行 `services.BroadcastLog = hub.Broadcast` 后生效 |
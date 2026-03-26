---
story_id: 1-6
story_key: NOKEY-1-6
title: WebSocket 实时日志推送
epic: epic-1
status: ready-for-dev
created: '2026-03-25'
---

# Story 1-6：WebSocket 实时日志推送

## 目标

实现编译日志的 WebSocket 实时推送，开发者在任务详情页可实时看到编译输出，后连接的客户端可回放历史日志。

---

## 依赖

| 前置 Story | 依赖内容 |
|------------|----------|
| Story 1-1 | `BuildTask` 模型（`LogPath`、`Status` 字段）|
| Story 1-4 | `QueueService.UpdateStatus()`、任务状态机 |
| Story 1-5 | `services.BroadcastLog` 函数变量（预留注入点）|

---

## 现有代码库上下文（必读）

### gorilla/websocket 已引入

**文件：** `controllers/layout.go`（第 3-21 行）

```go
import (
    "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true  // 内网部署，允许所有 origin
    },
}
```

**关键规则：**
- `upgrader` 已在 `layout.go` 中定义，**新的 `ws.go` 不要重复定义**，直接复用或在新文件中用相同包内访问
- 但 `ws.go` 和 `layout.go` 同属 `package controllers`，同包内 `upgrader` 直接可见，无需重新声明

### 现有 WS 骨架（不修改，仅参考结构）

**文件：** `controllers/layout.go`（第 82-102 行）

```go
func (this *LayoutController) Compilews() {
    Servews(this)
    this.EnableRender = false
}

func Servews(this *LayoutController) {
    conn, err := upgrader.Upgrade(this.Ctx.ResponseWriter, this.Ctx.Request, nil)
    if err != nil {
        panic(err)
    }
    err = conn.WriteMessage(websocket.TextMessage, []byte("hello"))
    // ...
}
```

**不要修改此函数。** 新 Story 在独立的 `controllers/ws.go` 实现。

### BroadcastLog 注入点（Story 1-5 预留）

**文件：** `services/compiler.go`（第 207 行）

```go
// 默认 noop，本 Story 绑定实际实现
var BroadcastLog func(taskId int64, line string) = func(int64, string) {}
```

本 Story 在 `main.go` 或 Hub 初始化时将此变量绑定为 `hub.Broadcast`。

### 路由注册方式

**文件：** `routers/router.go`

```go
func init() {
    beego.Router("/", &controllers.IndexController{})
    // ... 现有路由
}
```

WS 路由追加到 `init()` 末尾。

---

## 验收标准

- [ ] `/ws/build/:taskId` 建立 WebSocket 连接
- [ ] 后连接的客户端先收到历史日志文件内容（逐行发送，`type=log`）
- [ ] 编译进行中时新日志实时推送，延迟 ≤1 秒
- [ ] 编译完成后推送 `type=complete` 消息，客户端停止自动滚动
- [ ] 所有 WS 消息格式统一为 JSON：`{"type":"log|status|complete|error","taskId":123,"data":"...","timestamp":...}`
- [ ] 客户端断线重连：显示黄色提示条「连接已断开，正在重连...」，3 秒后自动重连
- [ ] Hub goroutine 管理所有连接，避免竞态；客户端断开自动 unregister

---

## 技术实现规范

### 文件结构

```
services/
└── ws_hub.go     ← 新建（Hub 核心）

controllers/
└── ws.go         ← 新建（WS 升级入口）

views/build/
└── detail.tpl    ← 新建（任务详情页，含 WS 客户端）

routers/router.go ← 修改，追加 WS 路由
main.go           ← 修改，初始化 Hub 并绑定 BroadcastLog
```

---

### `services/ws_hub.go` 实现规范

```go
package services

import (
    "bufio"
    "encoding/json"
    "fmt"
    "os"
    "sync"
    "time"

    "github.com/gorilla/websocket"
)

// WsMessage WebSocket 消息统一格式
type WsMessage struct {
    Type      string `json:"type"`       // log | status | complete | error
    TaskId    int64  `json:"taskId"`
    Data      string `json:"data"`
    Timestamp int64  `json:"timestamp"`
}

// client 代表一个 WebSocket 连接
type client struct {
    taskId int64
    send   chan []byte
    conn   *websocket.Conn
}

// Hub 管理所有 WebSocket 连接
type Hub struct {
    mu       sync.RWMutex
    clients  map[int64][]*client  // taskId → 连接列表
    register   chan *client
    unregister chan *client
}

var GlobalHub = NewHub()

func NewHub() *Hub {
    return &Hub{
        clients:    make(map[int64][]*client),
        register:   make(chan *client, 16),
        unregister: make(chan *client, 16),
    }
}

// Run 必须在独立 goroutine 中运行，管理注册/注销/广播
func (h *Hub) Run() {
    for {
        select {
        case c := <-h.register:
            h.mu.Lock()
            h.clients[c.taskId] = append(h.clients[c.taskId], c)
            h.mu.Unlock()

        case c := <-h.unregister:
            h.mu.Lock()
            list := h.clients[c.taskId]
            for i, cl := range list {
                if cl == c {
                    h.clients[c.taskId] = append(list[:i], list[i+1:]...)
                    close(c.send)
                    break
                }
            }
            h.mu.Unlock()
        }
    }
}

// Broadcast 向指定 taskId 的所有连接广播一行日志
// 此函数绑定到 services.BroadcastLog
func (h *Hub) Broadcast(taskId int64, line string) {
    msg := WsMessage{
        Type:      "log",
        TaskId:    taskId,
        Data:      line,
        Timestamp: time.Now().Unix(),
    }
    data, _ := json.Marshal(msg)

    h.mu.RLock()
    clients := h.clients[taskId]
    h.mu.RUnlock()

    for _, c := range clients {
        select {
        case c.send <- data:
        default:
            // 慢客户端：丢弃当前帧，避免阻塞
        }
    }
}

// BroadcastComplete 向指定 taskId 的所有连接发送编译完成消息
func (h *Hub) BroadcastComplete(taskId int64, status string) {
    msg := WsMessage{
        Type:      "complete",
        TaskId:    taskId,
        Data:      status,  // "success" 或 "failed"
        Timestamp: time.Now().Unix(),
    }
    data, _ := json.Marshal(msg)

    h.mu.RLock()
    clients := h.clients[taskId]
    h.mu.RUnlock()

    for _, c := range clients {
        select {
        case c.send <- data:
        default:
        }
    }
}

// ServeClient 为单个 WS 连接提供服务（历史日志回放 + 新消息转发）
// 在 controllers/ws.go 的 goroutine 中调用
func (h *Hub) ServeClient(c *client, logPath string) {
    // 1. 注册
    h.register <- c

    // 2. 启动写 goroutine（从 c.send channel 发往客户端）
    go func() {
        defer c.conn.Close()
        for data := range c.send {
            if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
                return
            }
        }
    }()

    // 3. 回放历史日志
    if logPath != "" {
        if f, err := os.Open(logPath); err == nil {
            scanner := bufio.NewScanner(f)
            for scanner.Scan() {
                msg := WsMessage{
                    Type:      "log",
                    TaskId:    c.taskId,
                    Data:      scanner.Text(),
                    Timestamp: time.Now().Unix(),
                }
                data, _ := json.Marshal(msg)
                select {
                case c.send <- data:
                default:
                }
            }
            f.Close()
        }
    }

    // 4. 阻塞读（客户端关闭时退出）
    for {
        _, _, err := c.conn.ReadMessage()
        if err != nil {
            break
        }
    }

    // 5. 注销
    h.unregister <- c
}

// BuildBroadcastFunc 返回可赋值给 services.BroadcastLog 的函数
func (h *Hub) BuildBroadcastFunc() func(int64, string) {
    return func(taskId int64, line string) {
        h.Broadcast(taskId, line)
    }
}
```

---

### `controllers/ws.go` 实现规范

```go
package controllers

import (
    "strconv"

    "gogs_tools/models"
    "gogs_tools/services"

    "github.com/astaxie/beego"
    "github.com/astaxie/beego/orm"
    "github.com/gorilla/websocket"
)

type WsController struct {
    beego.Controller  // 注意：不继承 BaseController，WS 不走 session 校验
}

// BuildLog GET /ws/build/:taskId
// Upgrade HTTP → WebSocket，回放历史日志后实时推送
func (this *WsController) BuildLog() {
    taskIdStr := this.Ctx.Input.Param(":taskId")
    taskId, err := strconv.ParseInt(taskIdStr, 10, 64)
    if err != nil || taskId <= 0 {
        this.Ctx.ResponseWriter.WriteHeader(400)
        return
    }

    // 查询任务（获取 LogPath）
    o := orm.NewOrm()
    task := models.BuildTask{Id: taskId}
    if err := o.Read(&task); err != nil {
        this.Ctx.ResponseWriter.WriteHeader(404)
        return
    }

    // Upgrade
    conn, err := upgrader.Upgrade(this.Ctx.ResponseWriter, this.Ctx.Request, nil)
    if err != nil {
        return
    }

    c := &services.ClientConn{
        TaskId: taskId,
        Send:   make(chan []byte, 256),
        Conn:   conn,
    }

    // 历史日志回放 + 实时推送（阻塞直到连接断开）
    services.GlobalHub.ServeClient(c, task.LogPath)
}
```

**注意：** `WsController` 继承 `beego.Controller` 而非 `BaseController`，避免 session 检查重定向破坏 WS 握手。

**`services.ClientConn` 导出结构体：** Hub 内部的 `client` 改为导出的 `ClientConn`，以便 controller 创建实例：

```go
// 在 ws_hub.go 中将内部 client 结构体改名为导出的 ClientConn：
type ClientConn struct {
    TaskId int64
    Send   chan []byte
    Conn   *websocket.Conn
}
```

---

### `main.go` 修改规范

在 `beego.Run()` 之前追加 Hub 启动并绑定 BroadcastLog：

```go
package main

import (
    _ "gogs_tools/routers"
    "gogs_tools/services"

    "github.com/astaxie/beego"
)

func main() {
    // 启动 Hub goroutine
    go services.GlobalHub.Run()

    // 绑定编译器广播钩子（Story 1-5 预留的注入点）
    services.BroadcastLog = services.GlobalHub.BuildBroadcastFunc()

    // 启动任务队列调度（Story 1-4）
    services.Recover()
    services.StartDispatcher()

    beego.Run()
}
```

**启动顺序：** Hub → BroadcastLog绑定 → Recover → StartDispatcher → beego.Run

---

### `routers/router.go` 修改规范

追加 WS 路由（注意用 `beego.Router` 而非 `AutoRouter`）：

```go
beego.Router("/ws/build/:taskId", &controllers.WsController{}, "get:BuildLog")
```

---

### `views/build/detail.tpl` 实现规范

继承主布局（`this.Layout = "index.tpl"`），内容区实现：

```html
<div class="row">
  <div class="col-lg-12">
    <!-- 状态提示条（断线重连）-->
    <div id="ws-alert" class="alert alert-warning" style="display:none;">
      连接已断开，正在重连...
    </div>

    <!-- 任务信息卡 -->
    <div class="ibox">
      <div class="ibox-title">
        <h5>任务 #{{.task.Id}} — {{.task.RepoName}}</h5>
        <span id="task-status" class="label label-{{.statusClass}}">{{.task.Status}}</span>
      </div>
      <div class="ibox-content">
        <p>Commit: <code>{{slice .task.CommitHash 0 7}}</code> &nbsp; 提交人: {{.task.Author}}</p>
        <p>{{.task.CommitMsg}}</p>
      </div>
    </div>

    <!-- 日志区 -->
    <div class="ibox">
      <div class="ibox-title"><h5>编译日志</h5></div>
      <div class="ibox-content">
        <pre id="log-box" style="height:500px;overflow-y:auto;background:#1a1a2e;color:#e0e0e0;padding:12px;"></pre>
      </div>
    </div>
  </div>
</div>

<script>
(function() {
    var taskId = {{.task.Id}};
    var logBox = document.getElementById('log-box');
    var wsAlert = document.getElementById('ws-alert');
    var autoScroll = true;
    var ws;
    var reconnectTimer;

    // 用户手动滚动时停止自动滚动
    logBox.addEventListener('scroll', function() {
        var atBottom = logBox.scrollHeight - logBox.scrollTop <= logBox.clientHeight + 20;
        autoScroll = atBottom;
    });

    function appendLine(text) {
        logBox.textContent += text + '\n';
        if (autoScroll) {
            logBox.scrollTop = logBox.scrollHeight;
        }
    }

    function connect() {
        var proto = location.protocol === 'https:' ? 'wss' : 'ws';
        ws = new WebSocket(proto + '://' + location.host + '/ws/build/' + taskId);

        ws.onopen = function() {
            wsAlert.style.display = 'none';
            clearTimeout(reconnectTimer);
        };

        ws.onmessage = function(e) {
            var msg = JSON.parse(e.data);
            if (msg.type === 'log') {
                appendLine(msg.data);
            } else if (msg.type === 'complete') {
                appendLine('[编译完成：' + msg.data + ']');
                autoScroll = false;  // 停止自动滚动
                ws.close();
            } else if (msg.type === 'error') {
                appendLine('[错误：' + msg.data + ']');
            }
        };

        ws.onerror = function() {
            wsAlert.style.display = 'block';
        };

        ws.onclose = function() {
            // 任务仍在运行时才重连
            wsAlert.style.display = 'block';
            reconnectTimer = setTimeout(connect, 3000);
        };
    }

    connect();
})();
</script>
```

---

### `controllers/build.go` 追加页面路由

Story 1-4 创建了 `BuildController`，本 Story 追加任务详情页面方法：

```go
// Detail GET /build/detail/:taskId — 渲染任务详情页
func (this *BuildController) Detail() {
    taskIdStr := this.Ctx.Input.Param(":taskId")
    taskId, err := strconv.ParseInt(taskIdStr, 10, 64)
    if err != nil || taskId <= 0 {
        this.Redirect("/build", 302)
        return
    }
    o := orm.NewOrm()
    task := models.BuildTask{Id: taskId}
    if err := o.Read(&task); err != nil {
        this.Redirect("/build", 302)
        return
    }
    // 状态→Bootstrap label class 映射
    statusClass := map[string]string{
        "pending": "default",
        "running": "warning",
        "success": "success",
        "failed":  "danger",
    }[task.Status]
    this.Data["task"] = task
    this.Data["statusClass"] = statusClass
    this.Data["menu"] = "build"
    this.Layout = "index.tpl"
    this.TplName = "build/detail.tpl"
}
```

同时在 `routers/router.go` 追加：

```go
beego.Router("/build/detail/:taskId", &controllers.BuildController{}, "get:Detail")
```

---

## 关键陷阱（LLM 常见错误）

| 陷阱 | 说明 |
|------|------|
| Hub 内部 map 并发读写 | 必须用 `sync.RWMutex`；Broadcast 用 `RLock`，register/unregister 用 `Lock` |
| WsController 不能继承 BaseController | session 检查会重定向，破坏 WS 握手。继承 `beego.Controller` |
| upgrader 重复定义 | `ws.go` 和 `layout.go` 同属 `package controllers`，直接复用已有 `upgrader` |
| 慢客户端阻塞广播 | send channel 用 buffered（256），写入失败用 `default` 丢弃当前帧 |
| BroadcastLog 绑定时机 | 必须在 `StartDispatcher()` 之前绑定，否则编译任务已开始但广播函数还是 noop |
| Beego WS 路由不支持 AutoRouter | WS 路由必须用 `beego.Router("/ws/build/:taskId", ...)` 显式注册 |
| Go 模板 `slice` 函数 | `{{slice .task.CommitHash 0 7}}` 取前7位；若 CommitHash 可能短于7位须先判断长度 |

---

## 验证清单

- [ ] `go build ./...` 无编译错误
- [ ] 访问 `/ws/build/1`（WebSocket 客户端），可收到 JSON 格式消息
- [ ] 手动触发任务后，浏览器日志区实时显示输出
- [ ] 关闭并重新连接，历史日志正确回放
- [ ] 服务端日志无 goroutine 泄漏警告
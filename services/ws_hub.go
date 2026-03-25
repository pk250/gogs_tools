package services

import (
	"bufio"
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WsMessage WebSocket 消息统一格式
type WsMessage struct {
	Type      string `json:"type"`      // log | status | complete | error
	TaskId    int64  `json:"taskId"`
	Data      string `json:"data"`
	Timestamp int64  `json:"timestamp"`
}

// ClientConn 代表一个 WebSocket 连接（导出以供 controllers 使用）
type ClientConn struct {
	TaskId int64
	Send   chan []byte
	Conn   *websocket.Conn
}

// Hub 管理所有 WebSocket 连接
type Hub struct {
	mu         sync.RWMutex
	clients    map[int64][]*ClientConn
	register   chan *ClientConn
	unregister chan *ClientConn
}

var GlobalHub = NewHub()

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[int64][]*ClientConn),
		register:   make(chan *ClientConn, 16),
		unregister: make(chan *ClientConn, 16),
	}
}

// Run 必须在独立 goroutine 中运行，管理注册/注销
func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			h.mu.Lock()
			h.clients[c.TaskId] = append(h.clients[c.TaskId], c)
			h.mu.Unlock()

		case c := <-h.unregister:
			h.mu.Lock()
			list := h.clients[c.TaskId]
			for i, cl := range list {
				if cl == c {
					h.clients[c.TaskId] = append(list[:i], list[i+1:]...)
					close(c.Send)
					break
				}
			}
			h.mu.Unlock()
		}
	}
}

// Broadcast 向指定 taskId 的所有连接广播一行日志
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
		case c.Send <- data:
		default:
			// 慢客户端：丢弃当前帧
		}
	}
}

// BroadcastComplete 向指定 taskId 的所有连接发送编译完成消息
func (h *Hub) BroadcastComplete(taskId int64, status string) {
	msg := WsMessage{
		Type:      "complete",
		TaskId:    taskId,
		Data:      status,
		Timestamp: time.Now().Unix(),
	}
	data, _ := json.Marshal(msg)

	h.mu.RLock()
	clients := h.clients[taskId]
	h.mu.RUnlock()

	for _, c := range clients {
		select {
		case c.Send <- data:
		default:
		}
	}
}

// BuildBroadcastFunc 返回可赋值给 services.BroadcastLog 的函数
func (h *Hub) BuildBroadcastFunc() func(int64, string) {
	return func(taskId int64, line string) {
		h.Broadcast(taskId, line)
	}
}

// ServeClient 为单个 WS 连接提供服务（历史日志回放 + 新消息转发）
func (h *Hub) ServeClient(c *ClientConn, logPath string) {
	// 1. 注册
	h.register <- c

	// 2. 启动写 goroutine
	go func() {
		defer c.Conn.Close()
		for data := range c.Send {
			if err := c.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
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
					TaskId:    c.TaskId,
					Data:      scanner.Text(),
					Timestamp: time.Now().Unix(),
				}
				data, _ := json.Marshal(msg)
				select {
				case c.Send <- data:
				default:
				}
			}
			f.Close()
		}
	}

	// 4. 阻塞读（客户端关闭时退出）
	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
	}

	// 5. 注销
	h.unregister <- c
}

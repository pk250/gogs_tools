package main

import (
	"context"
	_ "gogs_tools/routers"
	"gogs_tools/services"

	"github.com/astaxie/beego"
)

func main() {
	// 1. 启动 Hub goroutine
	go services.GlobalHub.Run()

	// 2. 绑定编译器广播钩子（必须在 StartDispatcher 之前）
	services.BroadcastLog = services.GlobalHub.BuildBroadcastFunc()

	// 3. 恢复上次未完成任务
	services.Recover()

	// 4. 启动 Dispatcher
	services.StartDispatcher(context.Background())

	// 5. 启动 HTTP 服务
	beego.Run()
}

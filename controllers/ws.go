package controllers

import (
	"strconv"

	"gogs_tools/models"
	"gogs_tools/services"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
)

type WsController struct {
	beego.Controller // 不继承 BaseController，避免 session 校验破坏 WS 握手
}

// BuildLog GET /ws/build/:taskId
func (this *WsController) BuildLog() {
	taskIdStr := this.Ctx.Input.Param(":taskId")
	taskId, err := strconv.ParseInt(taskIdStr, 10, 64)
	if err != nil || taskId <= 0 {
		this.Ctx.ResponseWriter.WriteHeader(400)
		return
	}

	o := orm.NewOrm()
	task := models.BuildTask{Id: taskId}
	if err := o.Read(&task); err != nil {
		this.Ctx.ResponseWriter.WriteHeader(404)
		return
	}

	conn, err := upgrader.Upgrade(this.Ctx.ResponseWriter, this.Ctx.Request, nil)
	if err != nil {
		return
	}

	c := &services.ClientConn{
		TaskId: taskId,
		Send:   make(chan []byte, 256),
		Conn:   conn,
	}

	services.GlobalHub.ServeClient(c, task.LogPath)
}

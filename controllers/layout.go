package controllers

import (
	"gogs_tools/models"
	"log"
	"net/http"
	"strconv"

	"github.com/astaxie/beego/orm"
	"github.com/gorilla/websocket"
)

type LayoutController struct {
	BaseController
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (this *LayoutController) Datainfo() {
	datainfos := make([]models.Datainfos, 0)
	o := orm.NewOrm()

	pages := this.Ctx.Input.Params()
	qs := o.QueryTable("datainfos")
	count, err := qs.Count()
	if err != nil {
		log.Println("err:", err)
	}
	this.Data["count"] = count
	if len(pages) > 0 {
		num, _ := strconv.ParseInt(pages["0"], 10, 64)
		_, err := qs.Limit(10, 10*(num-1)).All(&datainfos)
		if err != nil {
			log.Println("err:", err)
		} else {
			this.Data["datainfo"] = datainfos
			this.Data["pages"] = num
		}
	} else {
		_, err := qs.Limit(10, 0).All(&datainfos)
		if err != nil {
			log.Println("err:", err)
		} else {
			this.Data["datainfo"] = datainfos
			this.Data["pages"] = 1
		}
	}

	this.Data["menu"] = "datainfo"
	this.Layout = "index.tpl"
	this.TplName = "datainfo.tpl"
}

func (this *LayoutController) Compile() {
	this.Data["menu"] = "compile"
	this.Layout = "index.tpl"
	this.TplName = "compile.tpl"
}

func (this *LayoutController) Knowledge() {
	this.Data["menu"] = "knowledge"
	this.Layout = "index.tpl"
	this.TplName = "knowledge.tpl"
}

func (this *LayoutController) Messages() {
	this.Data["menu"] = "messages"
	this.Layout = "index.tpl"
	this.TplName = "knowledge.tpl"
}

func (this *LayoutController) Account() {
	this.Data["menu"] = "account"
	this.Layout = "index.tpl"
	this.TplName = "knowledge.tpl"
}

func (this *LayoutController) Compilews() {
	Servews(this)
	this.EnableRender = false
}

func Servews(this *LayoutController) {
	// hashCode := this.GetString("hashCode")

	conn, err := upgrader.Upgrade(this.Ctx.ResponseWriter, this.Ctx.Request, nil)
	if err != nil {
		panic(err)
	}

	// go func() {
	err = conn.WriteMessage(websocket.TextMessage, []byte("hello"))
	if err != nil {
		panic(err)
	}

	// }()
}

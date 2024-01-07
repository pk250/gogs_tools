package controllers

import (
	"gogs_tools/models"
	"log"

	"github.com/astaxie/beego/orm"
)

type LayoutController struct {
	BaseController
}

func (this *LayoutController) Datainfo() {
	datainfos := make([]models.Datainfos, 0)
	o := orm.NewOrm()

	_, err := o.Raw("SELECT * FROM datainfos ORDER BY id DESC LIMIT 100").QueryRows(&datainfos)

	if err != nil {
		log.Println("err:", err)
	} else {
		this.Data["datainfo"] = datainfos
	}

	this.Data["menu"] = "datainfo"
	this.Layout = "index.html"
	this.TplName = "datainfo.html"
}

func (this *LayoutController) Compile() {
	this.Data["menu"] = "compile"
	this.Layout = "index.html"
	this.TplName = "compile.html"
}

func (this *LayoutController) Knowledge() {
	this.Data["menu"] = "knowledge"
	this.Layout = "index.html"
	this.TplName = "knowledge.html"
}

func (this *LayoutController) Messages() {
	this.Data["menu"] = "messages"
	this.Layout = "index.html"
	this.TplName = "knowledge.html"
}

func (this *LayoutController) Account() {
	this.Data["menu"] = "account"
	this.Layout = "index.html"
	this.TplName = "knowledge.html"
}

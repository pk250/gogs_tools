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

	qs := o.QueryTable("datainfos")

	_, err := qs.Limit(10, 0).All(&datainfos)

	if err != nil {
		log.Println("err:", err)
	} else {
		this.Data["datainfo"] = datainfos
	}

	log.Println(datainfos)

	this.Layout = "index.html"
	this.TplName = "datainfo.html"
}

func (this *LayoutController) Compile() {
	this.Layout = "index.html"
	this.TplName = "compile.html"
}

func (this *LayoutController) Knowledge() {
	this.Layout = "index.html"
	this.TplName = "knowledge.html"
}

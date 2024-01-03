package controllers

import (
	"gogs_tools/models"
	"log"
	"strconv"

	"github.com/astaxie/beego/orm"
)

type LayoutController struct {
	BaseController
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

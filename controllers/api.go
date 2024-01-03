package controllers

import (
	"github.com/astaxie/beego"
)

type ApiController struct {
	beego.Controller
}

func (this *ApiController) Get() {
	this.Data["json"] = "api"
	this.ServeJSON()
}

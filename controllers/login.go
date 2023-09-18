package controllers

import (
	"gogs_tools/models"

	"github.com/astaxie/beego/orm"
)

type LoginController struct {
	BaseController
}

func (this *LoginController) Get() {
	this.TplName = "login.html"
}

func (this *LoginController) Post() {
	username := this.Input().Get("username")
	password := this.Input().Get("password")

	o := orm.NewOrm()
	user := models.Users{Username: username}
	err := o.Read(&user, "Username")
	if err != nil {
		this.Data["json"] = map[string]interface{}{"status": 0, "msg": "用户名或密码错误"}
		this.ServeJSON()
		return
	}

	if user.Password != password {
		this.Data["json"] = map[string]interface{}{"status": 0, "msg": "用户名或密码错误"}
		this.ServeJSON()
		return
	}

	this.SetSession("UserData", user)
	this.Data["json"] = map[string]interface{}{"status": 1, "msg": "登录成功", "url": "/"}
	this.ServeJSON()
	return
}

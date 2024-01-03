package controllers

import (
	"gogs_tools/models"
	"os/user"

	"github.com/astaxie/beego/orm"
)

type RegisterController struct {
	BaseController
}

func (this *RegisterController) Get() {
	this.TplName = "register.tpl"
}

func (this *RegisterController) Post() {
	username := this.Input().Get("username")
	password := this.Input().Get("password")
	email := this.Input().Get("email")

	if username == "" || password == "" || email == "" {
		this.Data["json"] = map[string]interface{}{"status": 0, "msg": "用户名、密码、邮箱不能为空"}
		this.ServeJSON()
		return
	}

	_, err := user.Lookup(username)
	if err == nil {
		this.Data["json"] = map[string]interface{}{"status": 0, "msg": "用户名已存在"}
		this.ServeJSON()
		return
	}

	o := orm.NewOrm()
	user := models.Users{Username: username, Password: password, Email: email}
	_, err = o.Insert(&user)
	if err != nil {
		this.Data["json"] = map[string]interface{}{"status": 0, "msg": "注册失败"}
		this.ServeJSON()
		return
	}

	this.Data["json"] = map[string]interface{}{"status": 1, "msg": "注册成功", "url": "/login"}
	this.ServeJSON()
	return
}

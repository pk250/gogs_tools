package controllers

import (
	"gogs_tools/models"

	"github.com/astaxie/beego/orm"
)

type RegisterController struct {
	BaseController
}

func (this *RegisterController) Get() {
	this.TplName = "register.html"
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
	o := orm.NewOrm()
	user := models.Users{Username: username}
	err := o.Read(&user, "Username")
	if err == orm.ErrNoRows {
		user.Email = email
		user.Password = password
		_, err = o.Insert(&user)
		if err != nil {
			this.Data["json"] = map[string]interface{}{"status": 0, "msg": "注册失败"}
			this.ServeJSON()
			return
		} else {
			this.SetSession("UserData", user)
			this.Data["json"] = map[string]interface{}{"status": 1, "msg": "注册成功", "url": "/"}
			this.ServeJSON()
			return
		}
	} else if err == nil {
		this.Data["json"] = map[string]interface{}{"status": 0, "msg": "用户名已存在"}
		this.ServeJSON()
		return
	}
}

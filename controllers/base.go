package controllers

import (
	"gogs_tools/models"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
)

type BaseController struct {
	beego.Controller
}

func (this *BaseController) Prepare() {
	users := this.Ctx.Input.Session("UserData")
	if users == nil && this.Ctx.Request.RequestURI != "/login" && this.Ctx.Request.RequestURI != "/register" {
		this.Ctx.Redirect(302, "/login")
	} else if users != nil {
		o := orm.NewOrm()
		user := models.Users{Id: users.(models.Users).Id}
		err := o.Read(&user)
		if err != nil {
			this.Ctx.Redirect(302, "/login")
		} else {
			if user.IsAdmin {
				this.Data["role"] = "管理员"
			} else {
				this.Data["role"] = "普通用户"
			}

			this.Data["Title"] = beego.AppConfig.String("Title")
			this.Data["username"] = user.Username
		}
	}
}

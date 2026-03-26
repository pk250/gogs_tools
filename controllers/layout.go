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
	o := orm.NewOrm()
	users := this.Ctx.Input.Session("UserData")
	if users == nil {
		this.Redirect("/login", 302)
		return
	}
	user := models.Users{Id: users.(models.Users).Id}
	o.Read(&user)
	this.Data["email"] = user.Email
	this.Data["menu"] = "account"
	this.Layout = "index.html"
	this.TplName = "account.tpl"
}

func (this *LayoutController) SaveAccount() {
	o := orm.NewOrm()
	users := this.Ctx.Input.Session("UserData")
	if users == nil {
		this.Data["json"] = map[string]interface{}{"code": 401, "message": "未登录"}
		this.ServeJSON()
		return
	}
	user := models.Users{Id: users.(models.Users).Id}
	if err := o.Read(&user); err != nil {
		this.Data["json"] = map[string]interface{}{"code": 404, "message": "用户不存在"}
		this.ServeJSON()
		return
	}
	email := this.GetString("email")
	newPwd := this.GetString("new_password")
	if email != "" {
		user.Email = email
	}
	if newPwd != "" {
		user.Password = newPwd
	}
	o.Update(&user, "Email", "Password")
	this.SetSession("UserData", user)
	this.Data["json"] = map[string]interface{}{"code": 0, "message": "保存成功"}
	this.ServeJSON()
}

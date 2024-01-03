package controllers

type LogoutController struct {
	BaseController
}

func (this *LogoutController) Get() {
	this.Ctx.Input.CruSession.Delete("UserData")
	this.TplName = "login.tpl"
}

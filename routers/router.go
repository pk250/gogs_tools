package routers

import (
	"gogs_tools/controllers"

	"github.com/astaxie/beego"
)

func init() {
	beego.Router("/", &controllers.IndexController{})
	beego.Router("/login", &controllers.LoginController{})
	beego.Router("/register", &controllers.RegisterController{})
	beego.Router("/logout", &controllers.LogoutController{})
	beego.AutoRouter(&controllers.ListsController{})
	beego.AutoRouter(&controllers.LayoutController{})
	beego.Router("/gogs", &controllers.GogsControllers{})
}

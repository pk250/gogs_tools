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
	beego.Router("/api", &controllers.ApiController{})
	beego.AutoRouter(&controllers.ListsController{})
	beego.AutoRouter(&controllers.LayoutController{})
	beego.Router("/gogs", &controllers.GogsControllers{})

	// Admin 路由（validate-path 必须在 /:id 前注册）
	beego.Router("/admin/keil-versions/validate-path", &controllers.AdminController{}, "post:KeilVersionValidatePath")
	beego.Router("/admin/keil-versions", &controllers.AdminController{}, "get:KeilVersionList;post:KeilVersionCreate")
	beego.Router("/admin/keil-versions/:page", &controllers.AdminController{}, "get:KeilVersionList")
	beego.Router("/admin/keil-versions/:id", &controllers.AdminController{}, "put:KeilVersionUpdate;delete:KeilVersionDelete")
	beego.Router("/admin/settings", &controllers.AdminController{}, "get:Settings;post:SaveSettings")
	beego.Router("/admin/team", &controllers.AdminController{}, "get:TeamView")

	// Repo 路由
	beego.Router("/repos", &controllers.RepoController{}, "get:List")
	beego.Router("/repos/:repoName/config", &controllers.RepoController{}, "get:Config;post:SaveConfig")
	beego.Router("/repos/:repoName/lint-config", &controllers.RepoController{}, "post:UploadLintConfig;delete:DeleteLintConfig")
	beego.Router("/repos/:repoName/lint-template", &controllers.RepoController{}, "get:LintTemplate")

	// Build 路由
	beego.Router("/api/build/trigger", &controllers.BuildController{}, "post:Trigger")
	beego.Router("/api/build/:taskId/enqueue", &controllers.BuildController{}, "post:EnqueueTask")
	beego.Router("/api/build/:taskId/webhook-retry", &controllers.BuildController{}, "post:WebhookRetry")

	// Build 详情页
	beego.Router("/build/detail/:taskId", &controllers.BuildController{}, "get:Detail")

	// 产物下载
	beego.Router("/build/artifacts/:taskId/:filename", &controllers.BuildController{}, "get:ArtifactDownload")

	// Dashboard
	beego.Router("/dashboard", &controllers.DashboardController{}, "get:Index")

	// WebSocket 路由
	beego.Router("/ws/build/:taskId", &controllers.WsController{}, "get:BuildLog")
}

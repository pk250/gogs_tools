package controllers

import (
	"gogs_tools/middleware"
	"gogs_tools/models"

	"github.com/astaxie/beego/orm"
)

type RepoController struct {
	BaseController
}

// List GET /repos
func (this *RepoController) List() {
	o := orm.NewOrm()

	var repoNames []orm.Params
	o.Raw("SELECT DISTINCT repository_name FROM gogs_d_b ORDER BY repository_name").Values(&repoNames)

	var configs []models.RepoConfig
	o.QueryTable("repo_config").All(&configs)

	configMap := make(map[string]models.RepoConfig)
	for _, c := range configs {
		configMap[c.RepoName] = c
	}

	var versions []models.KeilVersion
	o.QueryTable("keil_version").All(&versions)
	versionMap := make(map[int64]string)
	for _, v := range versions {
		versionMap[v.Id] = v.VersionName
	}

	this.Data["repoNames"] = repoNames
	this.Data["configMap"] = configMap
	this.Data["versionMap"] = versionMap
	this.Data["menu"] = "repos"
	this.Layout = "index.tpl"
	this.TplName = "repo/list.tpl"
}

// Config GET /repos/:repoName/config
func (this *RepoController) Config() {
	repoName := this.Ctx.Input.Param(":repoName")
	o := orm.NewOrm()

	config := models.RepoConfig{RepoName: repoName}
	err := o.Read(&config, "RepoName")
	if err == orm.ErrNoRows {
		config = models.RepoConfig{
			RepoName:     repoName,
			TriggerMode:  "manual",
			ArtifactName: repoName,
		}
	}

	var versions []models.KeilVersion
	o.QueryTable("keil_version").All(&versions)

	user := this.Ctx.Input.Session("UserData").(models.Users)
	this.Data["config"] = config
	this.Data["versions"] = versions
	this.Data["repoName"] = repoName
	this.Data["canEdit"] = middleware.CanEditRepoConfig(user)
	this.Data["menu"] = "repos"
	this.Layout = "index.tpl"
	this.TplName = "repo/config.tpl"
}

// SaveConfig POST /repos/:repoName/config
func (this *RepoController) SaveConfig() {
	user := this.Ctx.Input.Session("UserData").(models.Users)
	if !middleware.CanEditRepoConfig(user) {
		this.Ctx.ResponseWriter.WriteHeader(403)
		this.Data["json"] = map[string]interface{}{"code": 403, "message": "权限不足：当前为严格模式，仅管理员或项目负责人可修改配置"}
		this.ServeJSON()
		return
	}
	repoName := this.Ctx.Input.Param(":repoName")
	o := orm.NewOrm()

	keilVersionId, _ := this.GetInt64("keil_version_id")
	triggerMode := this.GetString("trigger_mode")
	artifactName := this.GetString("artifact_name")
	notifyEmails := this.GetString("notify_emails")
	webhookEnabled, _ := this.GetBool("webhook_enabled")
	webhookUrl := this.GetString("webhook_url")
	lintConfigPath := this.GetString("lint_config_path")

	if artifactName == "" {
		artifactName = repoName
	}

	config := models.RepoConfig{
		RepoName:       repoName,
		KeilVersionId:  keilVersionId,
		TriggerMode:    triggerMode,
		ArtifactName:   artifactName,
		NotifyEmails:   notifyEmails,
		WebhookEnabled: webhookEnabled,
		WebhookUrl:     webhookUrl,
		LintConfigPath: lintConfigPath,
	}

	exist := models.RepoConfig{RepoName: repoName}
	err := o.Read(&exist, "RepoName")
	if err == orm.ErrNoRows {
		_, err = o.Insert(&config)
	} else {
		config.Id = exist.Id
		_, err = o.Update(&config, "KeilVersionId", "TriggerMode", "ArtifactName",
			"NotifyEmails", "WebhookEnabled", "WebhookUrl", "LintConfigPath")
	}

	if err != nil {
		this.Data["json"] = map[string]interface{}{"code": 500, "message": "保存失败: " + err.Error()}
	} else {
		this.Data["json"] = map[string]interface{}{"code": 0, "message": "保存成功"}
	}
	this.ServeJSON()
}

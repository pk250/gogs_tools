package controllers

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

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
	this.Layout = "index.html"
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

	// lint config file info
	lintFileName := ""
	var lintUploadedAt time.Time
	if config.LintConfigPath != "" {
		if fi, err2 := os.Stat(config.LintConfigPath); err2 == nil {
			lintFileName = filepath.Base(config.LintConfigPath)
			lintUploadedAt = fi.ModTime()
		}
	}

	// default template path
	var tplCfg models.SysConfig
	lintTplURL := ""
	if err3 := o.QueryTable("sys_config").Filter("ConfigKey", models.ConfigKeyLintTplPath).One(&tplCfg); err3 == nil && tplCfg.ConfigVal != "" {
		lintTplURL = "/repos/" + repoName + "/lint-template"
	}

	user := this.Ctx.Input.Session("UserData").(models.Users)
	this.Data["config"] = config
	this.Data["versions"] = versions
	this.Data["repoName"] = repoName
	this.Data["canEdit"] = middleware.CanEditRepoConfig(user)
	this.Data["lintFileName"] = lintFileName
	this.Data["lintUploadedAt"] = lintUploadedAt
	this.Data["lintTplURL"] = lintTplURL
	this.Data["menu"] = "repos"
	this.Layout = "index.html"
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
	}

	exist := models.RepoConfig{RepoName: repoName}
	err := o.Read(&exist, "RepoName")
	if err == orm.ErrNoRows {
		_, err = o.Insert(&config)
	} else {
		config.Id = exist.Id
		_, err = o.Update(&config, "KeilVersionId", "TriggerMode", "ArtifactName",
			"NotifyEmails", "WebhookEnabled", "WebhookUrl")
	}

	if err != nil {
		this.Data["json"] = map[string]interface{}{"code": 500, "message": "保存失败: " + err.Error()}
	} else {
		this.Data["json"] = map[string]interface{}{"code": 0, "message": "保存成功"}
	}
	this.ServeJSON()
}

const lintConfigDir = "./data/lint-configs"
const maxLintSize = 1 * 1024 * 1024 // 1MB

// UploadLintConfig POST /repos/:repoName/lint-config
func (this *RepoController) UploadLintConfig() {
	user := this.Ctx.Input.Session("UserData").(models.Users)
	if !middleware.CanEditRepoConfig(user) {
		this.Ctx.ResponseWriter.WriteHeader(403)
		this.Data["json"] = map[string]interface{}{"code": 403, "message": "权限不足"}
		this.ServeJSON()
		return
	}
	repoName := this.Ctx.Input.Param(":repoName")
	f, h, err := this.GetFile("lint_file")
	if err != nil {
		this.Data["json"] = map[string]interface{}{"code": 400, "message": "读取文件失败"}
		this.ServeJSON()
		return
	}
	defer f.Close()
	if !strings.HasSuffix(strings.ToLower(h.Filename), ".lnt") {
		this.Data["json"] = map[string]interface{}{"code": 400, "message": "仅支持 .lnt 文件"}
		this.ServeJSON()
		return
	}
	if h.Size > maxLintSize {
		this.Data["json"] = map[string]interface{}{"code": 400, "message": "文件超过 1MB 限制"}
		this.ServeJSON()
		return
	}
	os.MkdirAll(lintConfigDir, 0755)
	safeRepo := filepath.Base(repoName)
	destPath := filepath.Join(lintConfigDir, fmt.Sprintf("%s.lnt", safeRepo))
	out, err := os.Create(destPath)
	if err != nil {
		this.Data["json"] = map[string]interface{}{"code": 500, "message": "保存文件失败"}
		this.ServeJSON()
		return
	}
	defer out.Close()
	io.Copy(out, f)
	// update config
	o := orm.NewOrm()
	config := models.RepoConfig{RepoName: repoName}
	if err2 := o.Read(&config, "RepoName"); err2 == orm.ErrNoRows {
		config = models.RepoConfig{RepoName: repoName, TriggerMode: "manual", ArtifactName: repoName}
		config.LintConfigPath = destPath
		o.Insert(&config)
	} else {
		config.LintConfigPath = destPath
		o.Update(&config, "LintConfigPath")
	}
	this.Data["json"] = map[string]interface{}{"code": 0, "message": "上传成功", "filename": h.Filename}
	this.ServeJSON()
}

// DeleteLintConfig DELETE /repos/:repoName/lint-config
func (this *RepoController) DeleteLintConfig() {
	user := this.Ctx.Input.Session("UserData").(models.Users)
	if !middleware.CanEditRepoConfig(user) {
		this.Ctx.ResponseWriter.WriteHeader(403)
		this.Data["json"] = map[string]interface{}{"code": 403, "message": "权限不足"}
		this.ServeJSON()
		return
	}
	repoName := this.Ctx.Input.Param(":repoName")
	o := orm.NewOrm()
	config := models.RepoConfig{RepoName: repoName}
	if err := o.Read(&config, "RepoName"); err != nil {
		this.Data["json"] = map[string]interface{}{"code": 404, "message": "仓库配置不存在"}
		this.ServeJSON()
		return
	}
	if config.LintConfigPath != "" {
		os.Remove(config.LintConfigPath)
	}
	config.LintConfigPath = ""
	o.Update(&config, "LintConfigPath")
	this.Data["json"] = map[string]interface{}{"code": 0, "message": "已删除"}
	this.ServeJSON()
}

// LintTemplate GET /repos/:repoName/lint-template
func (this *RepoController) LintTemplate() {
	o := orm.NewOrm()
	var tplCfg models.SysConfig
	if err := o.QueryTable("sys_config").Filter("ConfigKey", models.ConfigKeyLintTplPath).One(&tplCfg); err != nil || tplCfg.ConfigVal == "" {
		this.Ctx.ResponseWriter.WriteHeader(404)
		return
	}
	path := tplCfg.ConfigVal
	if _, err := os.Stat(path); err != nil {
		this.Ctx.ResponseWriter.WriteHeader(404)
		return
	}
	this.Ctx.Output.Download(path, filepath.Base(path))
}

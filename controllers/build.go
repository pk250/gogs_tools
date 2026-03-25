package controllers

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"gogs_tools/models"
	"gogs_tools/services"
	"gogs_tools/services/notifier"

	"github.com/astaxie/beego/orm"
)

type BuildController struct {
	BaseController
}

// Trigger POST /api/build/trigger
func (this *BuildController) Trigger() {
	type triggerReq struct {
		RepoName   string `json:"repo_name"`
		CommitHash string `json:"commit_hash"`
		CommitMsg  string `json:"commit_msg"`
		Author     string `json:"author"`
	}

	var req triggerReq
	if err := json.Unmarshal(this.Ctx.Input.RequestBody, &req); err != nil || req.RepoName == "" {
		this.Ctx.ResponseWriter.WriteHeader(400)
		this.Data["json"] = map[string]interface{}{"code": 400, "message": "参数错误"}
		this.ServeJSON()
		return
	}

	taskId, err := services.Enqueue(req.RepoName, req.CommitHash, req.CommitMsg, req.Author)
	if err != nil {
		if err.Error() == "queue full" {
			this.Ctx.ResponseWriter.WriteHeader(429)
			this.Data["json"] = map[string]interface{}{"code": 429, "message": "队列已满，请稍后重试"}
		} else {
			this.Ctx.ResponseWriter.WriteHeader(500)
			this.Data["json"] = map[string]interface{}{"code": 500, "message": "内部错误"}
		}
		this.ServeJSON()
		return
	}

	this.Data["json"] = map[string]interface{}{"code": 0, "data": map[string]interface{}{"task_id": taskId}}
	this.ServeJSON()
}

// Detail GET /build/detail/:taskId
func (this *BuildController) Detail() {
	taskIdStr := this.Ctx.Input.Param(":taskId")
	taskId, err := strconv.ParseInt(taskIdStr, 10, 64)
	if err != nil || taskId <= 0 {
		this.Redirect("/", 302)
		return
	}
	o := orm.NewOrm()
	task := models.BuildTask{Id: taskId}
	if err := o.Read(&task); err != nil {
		this.Redirect("/", 302)
		return
	}
	statusClass := map[string]string{
		"pending": "default",
		"running": "warning",
		"success": "success",
		"failed":  "danger",
	}[task.Status]

	// fetch repo config for trigger mode
	var cfg models.RepoConfig
	o.QueryTable("repo_config").Filter("RepoName", task.RepoName).One(&cfg)

	// list artifacts if present
	artifactsDir := filepath.Join(".", "data", "artifacts", taskIdStr)
	var artifacts []string
	if entries, err := ioutil.ReadDir(artifactsDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				artifacts = append(artifacts, e.Name())
			}
		}
	}

	// fetch lint result if present
	var lintResult models.ReviewResult
	hasLint := o.QueryTable("review_result").
		Filter("TaskId", taskId).
		Filter("ReviewType", models.ReviewTypeLint).
		One(&lintResult) == nil

	// fetch git check result if present
	var gitCheckResult models.ReviewResult
	hasGitCheck := o.QueryTable("review_result").
		Filter("TaskId", taskId).
		Filter("ReviewType", models.ReviewTypeGitMsg).
		One(&gitCheckResult) == nil

	// fetch AI review result if present
	var aiResult models.ReviewResult
	hasAI := o.QueryTable("review_result").
		Filter("TaskId", taskId).
		Filter("ReviewType", models.ReviewTypeAI).
		One(&aiResult) == nil

	this.Data["task"] = task
	this.Data["statusClass"] = statusClass
	this.Data["triggerMode"] = cfg.TriggerMode
	this.Data["hasWebhook"] = cfg.WebhookEnabled && cfg.WebhookUrl != ""
	this.Data["artifacts"] = artifacts
	this.Data["lintResult"] = lintResult
	this.Data["hasLint"] = hasLint
	this.Data["gitCheckResult"] = gitCheckResult
	this.Data["hasGitCheck"] = hasGitCheck
	this.Data["aiResult"] = aiResult
	this.Data["hasAI"] = hasAI
	this.Data["menu"] = "dashboard"
	this.Layout = "index.tpl"
	this.TplName = "build/detail.tpl"
}

// EnqueueTask POST /api/build/:taskId/enqueue — manually enqueue a pending task
func (this *BuildController) EnqueueTask() {
	taskIdStr := this.Ctx.Input.Param(":taskId")
	taskId, err := strconv.ParseInt(taskIdStr, 10, 64)
	if err != nil || taskId <= 0 {
		this.Data["json"] = map[string]interface{}{"code": 400, "message": "无效 taskId"}
		this.ServeJSON()
		return
	}
	o := orm.NewOrm()
	task := models.BuildTask{Id: taskId}
	if err := o.Read(&task); err != nil {
		this.Data["json"] = map[string]interface{}{"code": 404, "message": "任务不存在"}
		this.ServeJSON()
		return
	}
	if task.Status != models.TaskStatusPending {
		this.Data["json"] = map[string]interface{}{"code": 400, "message": "任务状态不是 pending"}
		this.ServeJSON()
		return
	}
	if err := services.EnqueueById(taskId); err != nil {
		this.Ctx.ResponseWriter.WriteHeader(429)
		this.Data["json"] = map[string]interface{}{"code": 429, "message": "队列已满，请稍后重试"}
		this.ServeJSON()
		return
	}
	this.Data["json"] = map[string]interface{}{"code": 0, "message": "ok"}
	this.ServeJSON()
}

// ArtifactDownload GET /build/artifacts/:taskId/:filename
func (this *BuildController) ArtifactDownload() {
	taskIdStr := this.Ctx.Input.Param(":taskId")
	filename := this.Ctx.Input.Param(":filename")
	// sanitise: no path separators allowed
	if filename == "" || filepath.Base(filename) != filename {
		this.Ctx.ResponseWriter.WriteHeader(400)
		return
	}
	path := filepath.Join(".", "data", "artifacts", taskIdStr, filename)
	if _, err := os.Stat(path); err != nil {
		this.Ctx.ResponseWriter.WriteHeader(404)
		return
	}
	this.Ctx.Output.Download(path, filename)
}

// WebhookRetry POST /api/build/:taskId/webhook-retry
func (this *BuildController) WebhookRetry() {
	taskIdStr := this.Ctx.Input.Param(":taskId")
	taskId, err := strconv.ParseInt(taskIdStr, 10, 64)
	if err != nil || taskId <= 0 {
		this.Data["json"] = map[string]interface{}{"code": 400, "message": "无效 taskId"}
		this.ServeJSON()
		return
	}
	o := orm.NewOrm()
	task := models.BuildTask{Id: taskId}
	if err := o.Read(&task); err != nil {
		this.Ctx.ResponseWriter.WriteHeader(404)
		this.Data["json"] = map[string]interface{}{"code": 404, "message": "任务不存在"}
		this.ServeJSON()
		return
	}
	go notifier.SendWebhook(task)
	this.Data["json"] = map[string]interface{}{"code": 0, "message": "webhook 回调已触发"}
	this.ServeJSON()
}

package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"gogs_tools/models"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

type webhookPayload struct {
	TaskId    int64    `json:"task_id"`
	RepoName  string   `json:"repo_name"`
	Status    string   `json:"status"`
	Commit    string   `json:"commit"`
	Author    string   `json:"author"`
	CommitMsg string   `json:"commit_msg"`
	Duration  string   `json:"duration"`
	DetailURL string   `json:"detail_url"`
	Artifacts []string `json:"artifacts"`
}

// SendWebhook fires the external webhook callback for a completed build task.
// Failures are logged but do not affect the main flow.
func SendWebhook(task models.BuildTask) {
	if err := sendWebhook(task); err != nil {
		logs.Warn("[Webhook] 回调失败 task=%d err=%v", task.Id, err)
		logs.Error("[Webhook] task=%d 回调失败记录: %s", task.Id, err.Error())
	}
}

func sendWebhook(task models.BuildTask) error {
	o := orm.NewOrm()
	var repoCfg models.RepoConfig
	if err := o.QueryTable("repo_config").Filter("RepoName", task.RepoName).One(&repoCfg); err != nil {
		return fmt.Errorf("读取仓库配置: %w", err)
	}
	if !repoCfg.WebhookEnabled || repoCfg.WebhookUrl == "" {
		return nil
	}

	var appCfg models.SysConfig
	baseURL := ""
	if err := o.QueryTable("sys_config").Filter("ConfigKey", models.ConfigKeyAppBaseURL).One(&appCfg); err == nil {
		baseURL = appCfg.ConfigVal
	}

	duration := ""
	if !task.StartedAt.IsZero() && !task.FinishedAt.IsZero() {
		duration = task.FinishedAt.Sub(task.StartedAt).Round(time.Second).String()
	}

	shortHash := task.CommitHash
	if len(shortHash) > 7 {
		shortHash = shortHash[:7]
	}

	payload := webhookPayload{
		TaskId:    task.Id,
		RepoName:  task.RepoName,
		Status:    task.Status,
		Commit:    shortHash,
		Author:    task.Author,
		CommitMsg: task.CommitMsg,
		Duration:  duration,
		DetailURL: fmt.Sprintf("%s/build/detail/%d", baseURL, task.Id),
		Artifacts: buildArtifactLinks(baseURL, task),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("json marshal: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(repoCfg.WebhookUrl, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("HTTP POST: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP %d from %s", resp.StatusCode, repoCfg.WebhookUrl)
	}
	return nil
}

// buildArtifactLinks returns download URLs for all artifacts of a task.
func buildArtifactLinks(baseURL string, task models.BuildTask) []string {
	dir := fmt.Sprintf("./data/artifacts/%d", task.Id)
	entries, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil
	}
	var links []string
	for _, e := range entries {
		if !e.IsDir() {
			links = append(links, fmt.Sprintf("%s/build/artifacts/%d/%s", baseURL, task.Id, e.Name()))
		}
	}
	return links
}

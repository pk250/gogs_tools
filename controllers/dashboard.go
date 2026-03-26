package controllers

import (
	"time"

	"gogs_tools/models"

	"github.com/astaxie/beego/orm"
)

type DashboardController struct {
	BaseController
}

// Index GET /dashboard
func (this *DashboardController) Index() {
	o := orm.NewOrm()
	const pageSize = 20

	pageNum, _ := this.GetInt64("page", 1)
	if pageNum < 1 {
		pageNum = 1
	}
	filterRepo := this.GetString("repo")
	filterStatus := this.GetString("status")
	filterAuthor := this.GetString("author")

	qs := o.QueryTable("build_task")
	if filterRepo != "" {
		qs = qs.Filter("RepoName", filterRepo)
	}
	if filterStatus != "" {
		qs = qs.Filter("Status", filterStatus)
	}
	if filterAuthor != "" {
		qs = qs.Filter("Author__icontains", filterAuthor)
	}

	count, _ := qs.Count()
	var tasks []models.BuildTask
	qs.OrderBy("-Id").Limit(pageSize, int(pageSize*(pageNum-1))).All(&tasks)

	totalPages := (count + int64(pageSize) - 1) / int64(pageSize)
	if totalPages == 0 {
		totalPages = 1
	}

	today := time.Now().Format("2006-01-02")
	var todaySuccess, todayFailed, todayRunning int64
	o.Raw("SELECT COUNT(*) FROM build_task WHERE status=? AND DATE(created_at)=?", models.TaskStatusSuccess, today).QueryRow(&todaySuccess)
	o.Raw("SELECT COUNT(*) FROM build_task WHERE status=? AND DATE(created_at)=?", models.TaskStatusFailed, today).QueryRow(&todayFailed)
	o.Raw("SELECT COUNT(*) FROM build_task WHERE status=?", models.TaskStatusRunning).QueryRow(&todayRunning)

	var repoRows []orm.Params
	o.Raw("SELECT DISTINCT repo_name FROM build_task ORDER BY repo_name").Values(&repoRows)
	var repoList []string
	for _, r := range repoRows {
		if name, ok := r["repo_name"].(string); ok {
			repoList = append(repoList, name)
		}
	}

	// 全部时间统计
	var totalSuccess, totalFailed int64
	o.Raw("SELECT COUNT(*) FROM build_task WHERE status=?", models.TaskStatusSuccess).QueryRow(&totalSuccess)
	o.Raw("SELECT COUNT(*) FROM build_task WHERE status=?", models.TaskStatusFailed).QueryRow(&totalFailed)

	// Top 5 仓库构建次数
	var topRepos []orm.Params
	o.Raw("SELECT repo_name, COUNT(*) as cnt FROM build_task GROUP BY repo_name ORDER BY cnt DESC LIMIT 5").Values(&topRepos)

	// 最近 7 天趋势
	type DayTrend struct {
		Day     string
		Success int64
		Failed  int64
	}
	var trends []DayTrend
	for i := 6; i >= 0; i-- {
		day := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		var s, f int64
		o.Raw("SELECT COUNT(*) FROM build_task WHERE status=? AND DATE(created_at)=?", models.TaskStatusSuccess, day).QueryRow(&s)
		o.Raw("SELECT COUNT(*) FROM build_task WHERE status=? AND DATE(created_at)=?", models.TaskStatusFailed, day).QueryRow(&f)
		trends = append(trends, DayTrend{Day: day[5:], Success: s, Failed: f})
	}

	this.Data["tasks"] = tasks
	this.Data["count"] = count
	this.Data["page"] = pageNum
	this.Data["totalPages"] = totalPages
	this.Data["prevPage"] = pageNum - 1
	this.Data["nextPage"] = pageNum + 1
	this.Data["filterRepo"] = filterRepo
	this.Data["filterStatus"] = filterStatus
	this.Data["filterAuthor"] = filterAuthor
	this.Data["repoList"] = repoList
	this.Data["todaySuccess"] = todaySuccess
	this.Data["todayFailed"] = todayFailed
	this.Data["todayRunning"] = todayRunning
	this.Data["totalSuccess"] = totalSuccess
	this.Data["totalFailed"] = totalFailed
	this.Data["topRepos"] = topRepos
	this.Data["trends"] = trends

	// build review status summary per task
	type reviewSummary struct {
		Label string
		Class string
	}
	taskReviewStatus := make(map[int64]reviewSummary)
	for _, t := range tasks {
		var results []models.ReviewResult
		o.QueryTable("review_result").Filter("TaskId", t.Id).All(&results)
		if len(results) == 0 {
			continue
		}
		hasFail, hasWarn := false, false
		for _, r := range results {
			if r.Status == models.ReviewStatusFail {
				hasFail = true
			} else if r.Status == models.ReviewStatusWarn {
				hasWarn = true
			}
		}
		if hasFail {
			taskReviewStatus[t.Id] = reviewSummary{Label: "有错误", Class: "danger"}
		} else if hasWarn {
			taskReviewStatus[t.Id] = reviewSummary{Label: "有警告", Class: "warning"}
		} else {
			taskReviewStatus[t.Id] = reviewSummary{Label: "通过", Class: "success"}
		}
	}
	this.Data["taskReviewStatus"] = taskReviewStatus
	this.Data["menu"] = "dashboard"
	this.Layout = "index.html"
	this.TplName = "dashboard/index.tpl"
}

package services

import (
	"fmt"
	"regexp"
	"time"

	"gogs_tools/models"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

// RunGitCheck 检查 commit message 是否符合管理员配置的正则规范
// 结果保存至 review_results 表（ReviewType = git_msg）
// 若未配置规范则跳过（status=skip）
func RunGitCheck(task models.BuildTask) {
	o := orm.NewOrm()

	// 1. 读取正则配置
	var cfgRow models.SysConfig
	err := o.QueryTable("sys_config").Filter("ConfigKey", models.ConfigKeyCommitMsgPattern).One(&cfgRow)
	if err != nil || cfgRow.ConfigVal == "" {
		saveGitCheckResult(o, task.Id, models.ReviewStatusSkip, "未配置 commit message 规范，跳过检查")
		return
	}
	pattern := cfgRow.ConfigVal

	// 2. 编译正则
	re, err := regexp.Compile(pattern)
	if err != nil {
		logs.Error("[GitCheck] 正则编译失败 pattern=%q err=%v", pattern, err)
		saveGitCheckResult(o, task.Id, models.ReviewStatusSkip, fmt.Sprintf("正则表达式无效: %v", err))
		return
	}

	// 3. 检查 commit message
	msg := task.CommitMsg
	if re.MatchString(msg) {
		saveGitCheckResult(o, task.Id, models.ReviewStatusPass,
			fmt.Sprintf("✅ commit message 符合规范: %q", truncate(msg, 80)))
	} else {
		saveGitCheckResult(o, task.Id, models.ReviewStatusFail,
			fmt.Sprintf("❌ commit message 不符合规范 (pattern: %s): %q", pattern, truncate(msg, 80)))
	}
}

func saveGitCheckResult(o orm.Ormer, taskId int64, status, summary string) {
	r := &models.ReviewResult{
		TaskId:     taskId,
		ReviewType: models.ReviewTypeGitMsg,
		Status:     status,
		Summary:    summary,
		CreatedAt:  time.Now(),
	}
	if _, err := o.Insert(r); err != nil {
		logs.Error("[GitCheck] saveGitCheckResult task=%d err=%v", taskId, err)
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

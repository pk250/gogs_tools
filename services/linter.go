package services

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"gogs_tools/models"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

// LintIssue 单条 Lint 问题
type LintIssue struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Code    string `json:"code"`
	Level   string `json:"level"` // warning | error | info
	Message string `json:"message"`
}

// RunLint 在编译成功后调用，执行 PC-Lint 并保存结果
// 若仓库未绑定 .lnt 文件则跳过（status=skip）
func RunLint(task models.BuildTask) {
	o := orm.NewOrm()

	// 1. 读取仓库配置
	repoConfig := models.RepoConfig{RepoName: task.RepoName}
	if err := o.Read(&repoConfig, "RepoName"); err != nil {
		saveLintResult(o, task.Id, models.ReviewStatusSkip, "无仓库配置", nil)
		return
	}

	// 2. 检查 .lnt 路径
	if repoConfig.LintConfigPath == "" {
		saveLintResult(o, task.Id, models.ReviewStatusSkip, "未上传 .lnt 配置文件，跳过 Lint", nil)
		return
	}

	// 3. 读取 PC-Lint 可执行路径
	var cfgRow models.SysConfig
	if err := o.QueryTable("sys_config").Filter("ConfigKey", models.ConfigKeyPCLintExe).One(&cfgRow); err != nil || cfgRow.ConfigVal == "" {
		saveLintResult(o, task.Id, models.ReviewStatusSkip, "未配置 PC-Lint 可执行路径", nil)
		return
	}
	pclintExe := cfgRow.ConfigVal

	// 4. 构造命令：pclint <config.lnt> <repo_dir/**/*.c>
	repoDir := repoConfig.ArtifactName
	if repoDir == "" {
		repoDir = task.RepoName
	}
	repoPath := fmt.Sprintf("%s/%s", ReposBaseDir, task.RepoName)

	cmd := exec.Command(pclintExe, repoConfig.LintConfigPath, "-u",
		fmt.Sprintf("+libdir(%s)", repoPath))
	cmd.Dir = repoPath

	out, err := cmd.CombinedOutput()
	outStr := string(out)
	logs.Info("[Lint] task=%d output_len=%d err=%v", task.Id, len(outStr), err)

	// 5. 解析输出
	issues := parseLintOutput(outStr)
	warnCount, errCount := 0, 0
	for _, iss := range issues {
		if iss.Level == "error" {
			errCount++
		} else {
			warnCount++
		}
	}

	status := models.ReviewStatusPass
	if errCount > 0 {
		status = models.ReviewStatusFail
	} else if warnCount > 0 {
		status = models.ReviewStatusWarn
	}
	summary := fmt.Sprintf("警告: %d，错误: %d", warnCount, errCount)
	saveLintResult(o, task.Id, status, summary, issues)
}

// parseLintOutput 解析 PC-Lint 标准输出
// 典型格式: file.c  123  error 123: message
var lintLineRe = regexp.MustCompile(`(?i)^(.+?)\s+(\d+)\s+(warning|error|info)\s+(\d+):\s+(.+)$`)

func parseLintOutput(output string) []LintIssue {
	var issues []LintIssue
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		m := lintLineRe.FindStringSubmatch(line)
		if len(m) == 6 {
			lineNum := 0
			fmt.Sscanf(m[2], "%d", &lineNum)
			issues = append(issues, LintIssue{
				File:    m[1],
				Line:    lineNum,
				Code:    m[4],
				Level:   strings.ToLower(m[3]),
				Message: m[5],
			})
		}
	}
	return issues
}

func saveLintResult(o orm.Ormer, taskId int64, status, summary string, issues []LintIssue) {
	detail := ""
	if issues != nil {
		if b, err := json.Marshal(issues); err == nil {
			detail = string(b)
		}
	}
	r := &models.ReviewResult{
		TaskId:     taskId,
		ReviewType: models.ReviewTypeLint,
		Status:     status,
		Summary:    summary,
		Detail:     detail,
		CreatedAt:  time.Now(),
	}
	if _, err := o.Insert(r); err != nil {
		logs.Error("[Lint] saveLintResult task=%d err=%v", taskId, err)
	}
}

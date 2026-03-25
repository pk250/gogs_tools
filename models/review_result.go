package models

import "time"

// ReviewResultStatus 审查结果状态
const (
	ReviewStatusPass = "pass"
	ReviewStatusWarn = "warn"
	ReviewStatusFail = "fail"
	ReviewStatusSkip = "skip"
)

// ReviewResultType 审查类型
const (
	ReviewTypeLint   = "lint"
	ReviewTypeAI     = "ai"
	ReviewTypeGitMsg = "git_msg"
)

type ReviewResult struct {
	Id         int64
	TaskId     int64     `orm:"index"`               // 关联 BuildTask.Id
	ReviewType string    `orm:"size(32)"`             // lint | ai | git_msg
	Status     string    `orm:"size(16)"`             // pass | warn | fail | skip
	Summary    string    `orm:"type(text);null"`      // 结果摘要
	Detail     string    `orm:"type(longtext);null"`  // 详细内容（JSON 或纯文本）
	CreatedAt  time.Time `orm:"auto_now_add;type(datetime)"`
}

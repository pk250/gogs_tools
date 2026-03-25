package models

import "time"

// BuildTaskStatus 任务状态枚举
const (
	TaskStatusPending = "pending"
	TaskStatusRunning = "running"
	TaskStatusSuccess = "success"
	TaskStatusFailed  = "failed"
)

type BuildTask struct {
	Id         int64
	RepoName   string    `orm:"size(128)"`                  // 仓库名
	CommitHash string    `orm:"size(64)"`                   // commit hash
	CommitMsg  string    `orm:"type(text);null"`            // commit message
	Author     string    `orm:"size(128);null"`             // 提交人
	Status     string    `orm:"size(16);default(pending)"`  // pending|running|success|failed
	LogPath    string    `orm:"size(512);null"`             // 日志文件路径
	ExitCode   int       `orm:"default(-1)"`                // 编译退出码
	StartedAt  time.Time `orm:"null;type(datetime)"`
	FinishedAt time.Time `orm:"null;type(datetime)"`
	CreatedAt  time.Time `orm:"auto_now_add;type(datetime)"`
	UpdatedAt  time.Time `orm:"auto_now;type(datetime)"`
}

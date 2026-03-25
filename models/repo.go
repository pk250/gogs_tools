package models

import "time"

// RepoConfig 仓库编译配置表
// 通过 RepoName 与 GogsDB.Repository_Name 关联（非外键，避免耦合）
type RepoConfig struct {
	Id             int64
	RepoName       string    `orm:"size(128);unique"`           // 仓库名，唯一
	KeilVersionId  int64     `orm:"default(0)"`                 // 关联 KeilVersion.Id
	TriggerMode    string    `orm:"size(16);default(manual)"`   // auto | manual
	ArtifactName   string    `orm:"size(128);null"`             // 产物文件名，默认仓库名
	NotifyEmails   string    `orm:"type(text);null"`            // 逗号分隔的邮件列表
	WebhookEnabled bool      `orm:"default(false)"`
	WebhookUrl     string    `orm:"size(512);null"`
	LintConfigPath string    `orm:"size(512);null"`             // .lnt 文件服务器路径
	CreatedAt      time.Time `orm:"auto_now_add;type(datetime)"`
	UpdatedAt      time.Time `orm:"auto_now;type(datetime)"`
}

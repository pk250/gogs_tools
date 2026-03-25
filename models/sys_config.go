package models

import "time"

// SysConfig 系统配置 KV 表
// 敏感值（AI Key、SMTP密码）存储时须加密，通过 ConfigService 读写
type SysConfig struct {
	Id        int64
	ConfigKey string    `orm:"size(128);unique"`
	ConfigVal string    `orm:"type(text);null"`   // 加密值或明文值
	IsSecret  bool      `orm:"default(false)"`    // true=加密存储
	UpdatedAt time.Time `orm:"auto_now;type(datetime)"`
}

// 预定义 ConfigKey 常量，避免硬编码字符串
const (
	ConfigKeyAIProvider   = "ai_provider"    // claude | openai
	ConfigKeyAIApiKey     = "ai_api_key"     // 加密存储
	ConfigKeyAIModel      = "ai_model"       // 模型名称
	ConfigKeyAIPrompt     = "ai_prompt"      // 审查提示词
	ConfigKeySMTPHost     = "smtp_host"
	ConfigKeySMTPPort     = "smtp_port"
	ConfigKeySMTPUser     = "smtp_user"
	ConfigKeySMTPPass     = "smtp_pass"      // 加密存储
	ConfigKeySMTPFrom     = "smtp_from"
	ConfigKeyReposBase    = "repos_base_path" // 仓库克隆根目录，默认 ./data/repos
	ConfigKeyAppBaseURL   = "app_base_url"   // 用于邮件中的详情页链接
	ConfigKeyPCLintExe         = "pclint_exe"          // PC-Lint 可执行文件完整路径
	ConfigKeyLintTplPath       = "lint_tpl_path"       // 默认 .lnt 模板文件路径
	ConfigKeyCommitMsgPattern  = "commit_msg_pattern"  // commit message 正则规范
)

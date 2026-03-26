package services

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gogs_tools/models"
	"gogs_tools/services/ai"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

// RunAIReview 在编译成功后调用，执行 AI 代码审查并保存结果
// 若未配置 AI 服务商或 API Key 则跳过（status=skip）
func RunAIReview(task models.BuildTask) {
	o := orm.NewOrm()

	// 1. 读取 AI 配置
	kv := loadAIConfig(o)
	provider := kv[models.ConfigKeyAIProvider]
	apiKeyEnc := kv[models.ConfigKeyAIApiKey]
	model := kv[models.ConfigKeyAIModel]
	prompt := kv[models.ConfigKeyAIPrompt]
	baseURL := kv[models.ConfigKeyAIBaseURL]

	if provider == "" || apiKeyEnc == "" {
		saveAIResult(o, task.Id, models.ReviewStatusSkip, "未配置 AI 服务商或 API Key", "")
		return
	}

	if prompt == "" {
		prompt = "请对以下 C/C++ 代码变更进行代码审查，指出潜在问题、安全隐患和改进建议，使用中文回答："
	}

	// 2. 解密 API Key
	apiKey, err := ai.Decrypt(apiKeyEnc)
	if err != nil {
		logs.Error("[AIReview] 解密 API Key 失败: %v", err)
		saveAIResult(o, task.Id, models.ReviewStatusSkip, "API Key 解密失败", "")
		return
	}

	// 3. 读取日志文件作为审查内容
	diff, err := readLogContent(task.LogPath)
	if err != nil || diff == "" {
		saveAIResult(o, task.Id, models.ReviewStatusSkip, "无编译日志内容可供审查", "")
		return
	}

	// 4. 调用 AI
	reviewer, err := ai.New(provider, apiKey, model, baseURL)
	if err != nil {
		logs.Error("[AIReview] 初始化 AI 适配器失败: %v", err)
		saveAIResult(o, task.Id, models.ReviewStatusSkip, fmt.Sprintf("AI 适配器初始化失败: %v", err), "")
		return
	}

	logs.Info("[AIReview] 开始审查 task=%d provider=%s", task.Id, provider)
	result, err := reviewer.Review(prompt, diff)
	if err != nil {
		logs.Error("[AIReview] task=%d 调用失败: %v", task.Id, err)
		saveAIResult(o, task.Id, models.ReviewStatusFail, fmt.Sprintf("AI 审查调用失败: %v", err), "")
		return
	}

	saveAIResult(o, task.Id, models.ReviewStatusPass, "AI 审查完成", result)
	logs.Info("[AIReview] task=%d 审查完成，结果长度=%d", task.Id, len(result))
}

func loadAIConfig(o orm.Ormer) map[string]string {
	keys := []string{
		models.ConfigKeyAIProvider,
		models.ConfigKeyAIApiKey,
		models.ConfigKeyAIModel,
		models.ConfigKeyAIPrompt,
		models.ConfigKeyAIBaseURL,
	}
	kv := make(map[string]string)
	for _, k := range keys {
		var c models.SysConfig
		if err := o.QueryTable("sys_config").Filter("ConfigKey", k).One(&c); err == nil {
			kv[k] = c.ConfigVal
		}
	}
	return kv
}

func readLogContent(logPath string) (string, error) {
	if logPath == "" {
		return "", nil
	}
	abs := logPath
	if !filepath.IsAbs(abs) {
		abs = filepath.Join(".", abs)
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		return "", err
	}
	s := string(data)
	// 限制最多 8000 字符，避免超出 token 限制
	if len(s) > 8000 {
		s = s[:8000] + "\n...(日志已截断)"
	}
	return s, nil
}

func saveAIResult(o orm.Ormer, taskId int64, status, summary, detail string) {
	r := &models.ReviewResult{
		TaskId:     taskId,
		ReviewType: models.ReviewTypeAI,
		Status:     status,
		Summary:    summary,
		Detail:     detail,
		CreatedAt:  time.Now(),
	}
	if _, err := o.Insert(r); err != nil {
		logs.Error("[AIReview] saveAIResult task=%d err=%v", taskId, err)
	}
}

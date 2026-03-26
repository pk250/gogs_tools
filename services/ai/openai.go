package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OpenAIReviewer 调用 OpenAI Chat Completions API
type OpenAIReviewer struct {
	APIKey  string
	Model   string
	BaseURL string // 自定义 Base URL，留空则使用官方地址
}

func (r *OpenAIReviewer) Review(prompt, diff string) (string, error) {
	model := r.Model
	if model == "" {
		model = "gpt-4o"
	}

	body := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt + "\n\n" + diff,
			},
		},
		"max_tokens": 1024,
	}

	bs, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	client := &http.Client{Timeout: 120 * time.Second}
	apiURL := "https://api.openai.com/v1/chat/completions"
	if r.BaseURL != "" {
		apiURL = strings.TrimRight(r.BaseURL, "/") + "/chat/completions"
	}
	req, err := http.NewRequest("POST", apiURL, bytes.NewReader(bs))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+r.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP 请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("OpenAI API 错误 %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("解析 OpenAI 响应失败: %w", err)
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("OpenAI 返回空内容")
	}
	return result.Choices[0].Message.Content, nil
}

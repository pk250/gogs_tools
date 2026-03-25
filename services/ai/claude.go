package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ClaudeReviewer 调用 Anthropic Claude API
type ClaudeReviewer struct {
	APIKey string
	Model  string
}

func (r *ClaudeReviewer) Review(prompt, diff string) (string, error) {
	model := r.Model
	if model == "" {
		model = "claude-sonnet-4-6"
	}

	body := map[string]interface{}{
		"model":      model,
		"max_tokens": 1024,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt + "\n\n" + diff,
			},
		},
	}

	bs, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	client := &http.Client{Timeout: 120 * time.Second}
	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(bs))
	if err != nil {
		return "", err
	}
	req.Header.Set("x-api-key", r.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("content-type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP 请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Claude API 错误 %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("解析 Claude 响应失败: %w", err)
	}
	if len(result.Content) == 0 {
		return "", fmt.Errorf("Claude 返回空内容")
	}
	return result.Content[0].Text, nil
}

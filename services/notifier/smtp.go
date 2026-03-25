package notifier

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/smtp"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gogs_tools/models"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

type smtpConfig struct {
	Host    string
	Port    int
	User    string
	Pass    string
	From    string
	BaseURL string
}

type attachment struct {
	name string
	data []byte
}

func loadConfig() (smtpConfig, error) {
	o := orm.NewOrm()
	keys := []string{
		models.ConfigKeySMTPHost,
		models.ConfigKeySMTPPort,
		models.ConfigKeySMTPUser,
		models.ConfigKeySMTPPass,
		models.ConfigKeySMTPFrom,
		models.ConfigKeyAppBaseURL,
	}
	kv := make(map[string]string)
	for _, k := range keys {
		var c models.SysConfig
		if err := o.QueryTable("sys_config").Filter("ConfigKey", k).One(&c); err == nil {
			kv[k] = c.ConfigVal
		}
	}
	if kv[models.ConfigKeySMTPHost] == "" {
		return smtpConfig{}, fmt.Errorf("SMTP 未配置")
	}
	port, _ := strconv.Atoi(kv[models.ConfigKeySMTPPort])
	if port == 0 {
		port = 587
	}
	cfg := smtpConfig{
		Host:    kv[models.ConfigKeySMTPHost],
		Port:    port,
		User:    kv[models.ConfigKeySMTPUser],
		Pass:    kv[models.ConfigKeySMTPPass],
		From:    kv[models.ConfigKeySMTPFrom],
		BaseURL: strings.TrimRight(kv[models.ConfigKeyAppBaseURL], "/"),
	}
	if cfg.From == "" {
		cfg.From = cfg.User
	}
	return cfg, nil
}

// SendBuildResult sends a notification email after compilation completes.
// Failures are logged but do not affect the main flow.
func SendBuildResult(task models.BuildTask) {
	if err := sendBuildResult(task); err != nil {
		logs.Warn("[Notifier] 发送邮件失败 task=%d err=%v", task.Id, err)
	}
}

func sendBuildResult(task models.BuildTask) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	o := orm.NewOrm()
	var repoCfg models.RepoConfig
	if err2 := o.QueryTable("repo_config").Filter("RepoName", task.RepoName).One(&repoCfg); err2 != nil {
		return fmt.Errorf("读取仓库配置: %w", err2)
	}
	if repoCfg.NotifyEmails == "" {
		return nil
	}
	to := splitEmails(repoCfg.NotifyEmails)
	if len(to) == 0 {
		return nil
	}

	shortHash := task.CommitHash
	if len(shortHash) > 7 {
		shortHash = shortHash[:7]
	}
	resultText := "成功"
	if task.Status == models.TaskStatusFailed {
		resultText = "失败"
	}
	subject := fmt.Sprintf("[gogs_tools] %s %s - 编译%s", task.RepoName, shortHash, resultText)

	duration := ""
	if !task.StartedAt.IsZero() && !task.FinishedAt.IsZero() {
		duration = task.FinishedAt.Sub(task.StartedAt).Round(time.Second).String()
	}
	detailURL := fmt.Sprintf("%s/build/detail/%d", cfg.BaseURL, task.Id)
	body := fmt.Sprintf(
		"编译状态：%s\n提交人：%s\n提交信息：%s\n编译耗时：%s\n详情页：%s\n",
		task.Status, task.Author, task.CommitMsg, duration, detailURL,
	)

	const maxAttachBytes int64 = 10 * 1024 * 1024
	var attachments []attachment

	if task.Status == models.TaskStatusSuccess {
		artDir := filepath.Join(".", "data", "artifacts", fmt.Sprintf("%d", task.Id))
		if entries, readErr := ioutil.ReadDir(artDir); readErr == nil {
			for _, e := range entries {
				if e.IsDir() {
					continue
				}
				ext := strings.ToLower(filepath.Ext(e.Name()))
				if ext != ".hex" && ext != ".bin" {
					continue
				}
				if e.Size() > maxAttachBytes {
					body += fmt.Sprintf("\n产物 %s 超过 10MB，请前往详情页下载。", e.Name())
					continue
				}
				if data, rErr := ioutil.ReadFile(filepath.Join(artDir, e.Name())); rErr == nil {
					attachments = append(attachments, attachment{e.Name(), data})
				}
			}
		}
	} else if task.Status == models.TaskStatusFailed && task.LogPath != "" {
		if fi, statErr := os.Stat(task.LogPath); statErr == nil && fi.Size() <= maxAttachBytes {
			if data, rErr := ioutil.ReadFile(task.LogPath); rErr == nil {
				attachments = append(attachments, attachment{filepath.Base(task.LogPath), data})
			}
		}
	}

	msgBytes := buildMIME(cfg.From, to, subject, body, attachments)
	return sendSMTP(cfg, to, msgBytes)
}

func buildMIME(from string, to []string, subject, body string, attachments []attachment) []byte {
	var buf bytes.Buffer
	encSubject := mime.QEncoding.Encode("utf-8", subject)
	buf.WriteString("From: " + from + "\r\n")
	buf.WriteString("To: " + strings.Join(to, ", ") + "\r\n")
	buf.WriteString("Subject: " + encSubject + "\r\n")
	buf.WriteString("MIME-Version: 1.0\r\n")

	if len(attachments) == 0 {
		buf.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
		buf.WriteString("\r\n")
		buf.WriteString(body)
		return buf.Bytes()
	}

	mw := multipart.NewWriter(&buf)
	buf.WriteString("Content-Type: multipart/mixed; boundary=\"" + mw.Boundary() + "\"\r\n")
	buf.WriteString("\r\n")

	ph := make(map[string][]string)
	ph["Content-Type"] = []string{"text/plain; charset=utf-8"}
	part, _ := mw.CreatePart(ph)
	part.Write([]byte(body))

	for _, a := range attachments {
		ah := make(map[string][]string)
		ah["Content-Type"] = []string{"application/octet-stream"}
		ah["Content-Transfer-Encoding"] = []string{"base64"}
		ah["Content-Disposition"] = []string{fmt.Sprintf("attachment; filename=\"%s\"", a.name)}
		ap, _ := mw.CreatePart(ah)
		encoded := base64.StdEncoding.EncodeToString(a.data)
		for i := 0; i < len(encoded); i += 76 {
			end := i + 76
			if end > len(encoded) {
				end = len(encoded)
			}
			ap.Write([]byte(encoded[i:end] + "\r\n"))
		}
	}
	mw.Close()
	return buf.Bytes()
}

func sendSMTP(cfg smtpConfig, to []string, msg []byte) error {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	auth := smtp.PlainAuth("", cfg.User, cfg.Pass, cfg.Host)

	if cfg.Port == 465 {
		tlsCfg := &tls.Config{ServerName: cfg.Host}
		conn, err := tls.Dial("tcp", addr, tlsCfg)
		if err != nil {
			return fmt.Errorf("TLS dial: %w", err)
		}
		defer conn.Close()
		client, err := smtp.NewClient(conn, cfg.Host)
		if err != nil {
			return fmt.Errorf("smtp client: %w", err)
		}
		defer client.Close()
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
		if err := client.Mail(cfg.From); err != nil {
			return err
		}
		for _, r := range to {
			if err := client.Rcpt(r); err != nil {
				return err
			}
		}
		w, err := client.Data()
		if err != nil {
			return err
		}
		defer w.Close()
		_, err = w.Write(msg)
		return err
	}

	return smtp.SendMail(addr, auth, cfg.From, to, msg)
}

func splitEmails(s string) []string {
	var result []string
	for _, e := range strings.Split(s, ",") {
		e = strings.TrimSpace(e)
		if e != "" && strings.Contains(e, "@") {
			result = append(result, e)
		}
	}
	return result
}

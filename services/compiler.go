package services

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"gogs_tools/models"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

const (
	CompileTimeout = 90 * time.Minute
	LogsBaseDir    = "./data/logs"
	ArtifactsDir   = "./data/artifacts"
	ReposBaseDir   = "./data/repos"
	ArtifactExts   = ".axf .hex .bin .map"
)

func init() {
	os.MkdirAll(LogsBaseDir, 0755)
	os.MkdirAll(ArtifactsDir, 0755)
	os.MkdirAll(ReposBaseDir, 0755)
}

// BroadcastLog 可注入的 WebSocket 广播钩子，默认 noop，Story 1-6 绑定实现
var BroadcastLog func(taskId int64, line string) = func(int64, string) {}

// reposBaseDir 返回仓库根目录，优先读取 SysConfig，回退到默认值
func reposBaseDir(o orm.Ormer) string {
	var c models.SysConfig
	if err := o.QueryTable("sys_config").Filter("ConfigKey", models.ConfigKeyReposBase).One(&c); err == nil && c.ConfigVal != "" {
		return c.ConfigVal
	}
	return ReposBaseDir
}

// Run 执行 Keil 编译流程，返回 nil 表示编译成功（退出码 0）
func Run(task models.BuildTask) error {
	o := orm.NewOrm()

	// 1. 读取仓库配置
	repoConfig := models.RepoConfig{RepoName: task.RepoName}
	if err := o.Read(&repoConfig, "RepoName"); err != nil {
		return fmt.Errorf("读取仓库配置失败: %w", err)
	}

	// 2. 读取 Keil 版本
	keilVersion := models.KeilVersion{Id: repoConfig.KeilVersionId}
	if err := o.Read(&keilVersion); err != nil {
		return fmt.Errorf("读取 Keil 版本失败 (id=%d): %w", repoConfig.KeilVersionId, err)
	}

	// 3. 查找 uvprojx 文件
	repoDir := filepath.Join(reposBaseDir(o), task.RepoName)
	uvprojxPath, err := findUvprojx(repoDir)
	if err != nil {
		return err
	}

	// 4. 准备日志文件
	if err := os.MkdirAll(LogsBaseDir, 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %w", err)
	}
	logPath := filepath.Join(LogsBaseDir, fmt.Sprintf("%d.log", task.Id))
	logFile, err := os.Create(logPath)
	if err != nil {
		return fmt.Errorf("创建日志文件失败: %w", err)
	}
	defer logFile.Close()

	// 5. 更新日志路径到 DB
	o.QueryTable("build_task").Filter("Id", task.Id).Update(orm.Params{"LogPath": logPath})

	// 6. 执行编译（含超时）
	ctx, cancel := context.WithTimeout(context.Background(), CompileTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, keilVersion.Uv4Path, "-rebuild", uvprojxPath, "-j0")

	// 用 io.Pipe 合并 stdout+stderr
	pr, pw := io.Pipe()
	cmd.Stdout = pw
	cmd.Stderr = pw

	if err := cmd.Start(); err != nil {
		pw.Close()
		return fmt.Errorf("启动 UV4.exe 失败: %w", err)
	}

	var cmdErr error
	go func() {
		cmdErr = cmd.Wait()
		pw.Close()
	}()

	// 7. 流式读取日志
	var last20Lines []string
	scanner := bufio.NewScanner(pr)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Fprintln(logFile, line)
		BroadcastLog(task.Id, line)
		last20Lines = append(last20Lines, line)
		if len(last20Lines) > 20 {
			last20Lines = last20Lines[1:]
		}
	}

	// 8. 处理超时
	if ctx.Err() == context.DeadlineExceeded {
		timeoutMsg := "[gogs_tools] 编译超时（90分钟），已强制终止"
		fmt.Fprintln(logFile, timeoutMsg)
		BroadcastLog(task.Id, timeoutMsg)
		o.QueryTable("build_task").Filter("Id", task.Id).Update(orm.Params{
			"error_summary": timeoutMsg,
			"finished_at":   time.Now(),
		})
		return fmt.Errorf("编译超时")
	}

	// 9. 更新 ErrorSummary（最后 20 行）
	summary := strings.Join(last20Lines, "\n")
	o.QueryTable("build_task").Filter("Id", task.Id).Update(orm.Params{
		"error_summary": summary,
		"finished_at":   time.Now(),
	})

	exitCode := 0
	if cmd.ProcessState != nil {
		exitCode = cmd.ProcessState.ExitCode()
	}
	o.QueryTable("build_task").Filter("Id", task.Id).Update(orm.Params{"exit_code": exitCode})

	if cmdErr != nil {
		logs.Error("[compiler] 任务 %d 编译失败，退出码非0: %v", task.Id, cmdErr)
		return fmt.Errorf("编译失败: %w", cmdErr)
	}

	// 10. 复制产物
	if err := copyArtifacts(task, repoDir, repoConfig.ArtifactName); err != nil {
		logs.Warn("[compiler] 任务 %d 复制产物失败（不影响状态）: %v", task.Id, err)
	}

	return nil
}

// copyArtifacts 复制编译产物到 /data/artifacts/{taskId}/
// 若 artifactName 非空且只找到一个主产物，则按配置名重命名（保留原扩展名）
func copyArtifacts(task models.BuildTask, repoDir, artifactName string) error {
	destDir := filepath.Join(ArtifactsDir, fmt.Sprintf("%d", task.Id))
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("创建产物目录失败: %w", err)
	}
	exts := strings.Fields(ArtifactExts)
	type srcFile struct{ path, name string }
	var files []srcFile
	filepath.Walk(repoDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		for _, e := range exts {
			if ext == e {
				files = append(files, srcFile{path, info.Name()})
			}
		}
		return nil
	})
	if len(files) == 0 {
		return fmt.Errorf("未找到任何产物文件（.axf/.hex/.bin/.map）")
	}
	for i, f := range files {
		destName := f.name
		// 单个产物且配置了 ArtifactName 时按配置重命名
		if len(files) == 1 && artifactName != "" {
			ext := filepath.Ext(f.name)
			base := strings.TrimSuffix(artifactName, ext)
			destName = base + ext
		}
		_ = i
		if copyErr := copyFile(f.path, filepath.Join(destDir, destName)); copyErr != nil {
			logs.Warn("[compiler] 复制产物 %s 失败: %v", f.name, copyErr)
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

func findUvprojx(repoDir string) (string, error) {
	var found string
	filepath.Walk(repoDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || found != "" {
			return nil
		}
		if strings.HasSuffix(strings.ToLower(info.Name()), ".uvprojx") {
			found = path
		}
		return nil
	})
	if found == "" {
		return "", fmt.Errorf("仓库 %s 中未找到 .uvprojx 文件", repoDir)
	}
	return found, nil
}

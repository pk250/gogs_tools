package services

import (
	"fmt"
	"time"

	"gogs_tools/models"

	"github.com/astaxie/beego/orm"
)

const (
	MaxWorkers  = 5
	MaxQueueCap = 50
)

// TaskCh 是全局任务 channel，Dispatcher 从中消费
var TaskCh = make(chan int64, MaxQueueCap)

// EnqueueById 将已存在的 pending 任务 ID 推入队列（用于手动触发）
func EnqueueById(taskId int64) error {
	if len(TaskCh) >= MaxQueueCap {
		return fmt.Errorf("queue full")
	}
	select {
	case TaskCh <- taskId:
	default:
		return fmt.Errorf("queue full")
	}
	return nil
}

// Enqueue 创建任务并入队，返回 task ID
// 若 channel 已满，返回 error（调用方返回 HTTP 429）
func Enqueue(repoName, commitHash, commitMsg, author string) (int64, error) {
	if len(TaskCh) >= MaxQueueCap {
		return 0, fmt.Errorf("queue full")
	}

	o := orm.NewOrm()
	task := &models.BuildTask{
		RepoName:   repoName,
		CommitHash: commitHash,
		CommitMsg:  commitMsg,
		Author:     author,
		Status:     models.TaskStatusPending,
	}
	id, err := o.Insert(task)
	if err != nil {
		return 0, fmt.Errorf("insert build_task: %w", err)
	}

	select {
	case TaskCh <- id:
	default:
		// channel 刚好在检查后满了，任务已在 DB，Recover 会处理
	}

	return id, nil
}

// UpdateStatus 是任务状态变更的唯一入口
func UpdateStatus(taskId int64, status string) error {
	o := orm.NewOrm()
	task := &models.BuildTask{Id: taskId}
	if err := o.Read(task); err != nil {
		return fmt.Errorf("read task %d: %w", taskId, err)
	}
	task.Status = status
	if status == models.TaskStatusRunning {
		task.StartedAt = time.Now()
	} else if status == models.TaskStatusSuccess || status == models.TaskStatusFailed {
		task.FinishedAt = time.Now()
	}
	_, err := o.Update(task)
	return err
}

// Recover 在服务启动时调用，将 running→pending，并重新入队 pending 任务
func Recover() {
	o := orm.NewOrm()

	// running → pending（上次崩溃未完成的任务）
	o.Raw("UPDATE build_task SET status=? WHERE status=?",
		models.TaskStatusPending, models.TaskStatusRunning).Exec()

	// 将所有 pending 任务重新入队（按 ID 排序）
	var tasks []*models.BuildTask
	o.QueryTable("build_task").
		Filter("Status", models.TaskStatusPending).
		OrderBy("Id").
		Limit(MaxQueueCap).
		All(&tasks)

	for _, t := range tasks {
		select {
		case TaskCh <- t.Id:
		default:
		}
	}
}

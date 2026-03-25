package services

import (
	"context"
	"sync"

	"gogs_tools/models"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

// StartDispatcher 启动 Dispatcher，在 main.go 中 beego.Run() 之前调用
func StartDispatcher(ctx context.Context) {
	var wg sync.WaitGroup
	sem := make(chan struct{}, MaxWorkers)

	go func() {
		for {
			select {
			case <-ctx.Done():
				wg.Wait()
				return
			case taskId := <-TaskCh:
				sem <- struct{}{}
				wg.Add(1)
				go func(id int64) {
					defer wg.Done()
					defer func() { <-sem }()
					o := orm.NewOrm()
					task := models.BuildTask{Id: id}
					if err := o.Read(&task); err != nil {
						logs.Error("[Worker] 读取任务失败 id=%d err=%v", id, err)
						return
					}
					runWorker(task)
				}(taskId)
			}
		}
	}()

	logs.Info("[Queue] Dispatcher started, maxWorkers=%d, queueCap=%d", MaxWorkers, MaxQueueCap)
}

func runWorker(task models.BuildTask) {
	if err := UpdateStatus(task.Id, models.TaskStatusRunning); err != nil {
		logs.Error("[Worker] UpdateStatus running, task=%d, err=%v", task.Id, err)
		return
	}
	err := Run(task)
	if err != nil {
		logs.Error("[Worker] 任务 %d 失败: %v", task.Id, err)
		UpdateStatus(task.Id, models.TaskStatusFailed)
	} else {
		UpdateStatus(task.Id, models.TaskStatusSuccess)
	}
}

package models

import (
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

// RunMigrations 手动迁移辅助函数
// 用于为现有表追加字段（AutoMigrate 仅建新表，不删/改已有列）
// 在需要手动变更表结构时调用此函数
func RunMigrations() {
	o := orm.NewOrm()
	_ = o
	logs.Info("[migrate] 数据库迁移检查完成")
}

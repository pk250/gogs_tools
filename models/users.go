package models

import (
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
)

type Users struct {
	Id         int64
	Username   string    `orm:"size(64);unique"`
	Password   string    `orm:"size(64)"`
	Email      string    `orm:"size(64)"`
	LastLogin  time.Time `orm:"auto_now_add;type(datetime)"`
	LoginCount int64     `orm:"default(0)"`
	IsActive   bool      `orm:"default(true)"`
	IsAdmin    bool      `orm:"default(false)"`
	IsStaff    bool      `orm:"default(false)"`
	Created    time.Time `orm:"auto_now_add;type(datetime)"`
	Updated    time.Time `orm:"auto_now;type(datetime)"`
}

func init() {
	orm.RegisterDataBase("default", "mysql", beego.AppConfig.String("sqlconn"))

	orm.RegisterModel(new(Datainfos))

	orm.RegisterModel(new(Users))

	orm.RegisterModel(new(GogsDB))

	orm.RunSyncdb("default", false, true)
}

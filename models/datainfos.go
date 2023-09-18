package models

import (
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Datainfos struct {
	Id            int64
	StorageName   string    `orm:"size(128);unique"`
	CommitValue   string    `orm:"size(32)"`
	CommitTime    string    `orm:"size(64)"`
	CommitAuth    string    `orm:"size(64)"`
	CompileStatus bool      `orm:"default(false)"`
	CompileUser   string    `orm:"size(64)"`
	CompileTime   time.Time `orm:"type(datetime)"`
	CompilePath   string    `orm:"size(128)"`
}

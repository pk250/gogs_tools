package models

import (
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Datainfos struct {
	Id            int64
	StorageName   string `orm:"size(128)"`
	CommitValue   string `orm:"size(64);unique"`
	CommitTime    time.Time
	CommitAuth    string    `orm:"size(64)"`
	CompileStatus bool      `orm:"default(false)"`
	CompileUser   string    `orm:"size(64);null"`
	CompileTime   time.Time `orm:"null"`
	CompilePath   string    `orm:"size(128);null"`
}

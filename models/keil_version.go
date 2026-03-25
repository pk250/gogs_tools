package models

import "time"

type KeilVersion struct {
	Id          int64
	VersionName string    `orm:"size(64);unique"`
	Uv4Path     string    `orm:"size(512)"`
	CreatedAt   time.Time `orm:"auto_now_add;type(datetime)"`
	UpdatedAt   time.Time `orm:"auto_now;type(datetime)"`
}

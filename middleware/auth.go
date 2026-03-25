package middleware

import (
	"gogs_tools/models"

	"github.com/astaxie/beego/orm"
)

// GetPermissionMode returns the current permission mode ("loose" or "strict").
// Defaults to "loose" if not configured.
func GetPermissionMode() string {
	o := orm.NewOrm()
	var c models.SysConfig
	if err := o.QueryTable("sys_config").Filter("ConfigKey", models.ConfigKeyPermissionMode).One(&c); err == nil {
		if c.ConfigVal == "strict" {
			return "strict"
		}
	}
	return "loose"
}

// CanEditRepoConfig returns true if the given user is allowed to edit repo
// configuration under the current permission mode.
// loose  → any authenticated user may edit
// strict → only admin or project_lead
func CanEditRepoConfig(user models.Users) bool {
	if user.IsAdmin || user.IsStaff {
		return true
	}
	if GetPermissionMode() == "loose" {
		return true
	}
	return false
}

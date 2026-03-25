package controllers

import (
	"gogs_tools/models"
	aiPkg "gogs_tools/services/ai"
	"os"
	"strconv"

	"github.com/astaxie/beego/orm"
)

type AdminController struct {
	BaseController
}

func (this *AdminController) Prepare() {
	this.BaseController.Prepare()
	users := this.Ctx.Input.Session("UserData")
	if users != nil {
		o := orm.NewOrm()
		user := models.Users{Id: users.(models.Users).Id}
		o.Read(&user)
		if !user.IsAdmin {
			this.Redirect("/layout/datainfo", 302)
			return
		}
	}
}

// KeilVersionList GET /admin/keil-versions 或 /admin/keil-versions/:page
func (this *AdminController) KeilVersionList() {
	o := orm.NewOrm()
	const pageSize = 10

	pages := this.Ctx.Input.Params()
	var pageNum int64 = 1
	if p, ok := pages[":page"]; ok && p != "" {
		if n, err := strconv.ParseInt(p, 10, 64); err == nil && n > 0 {
			pageNum = n
		}
	}

	qs := o.QueryTable("keil_version")
	count, _ := qs.Count()

	versions := make([]models.KeilVersion, 0)
	qs.OrderBy("-Id").Limit(pageSize, pageSize*(pageNum-1)).All(&versions)

	totalPages := (count + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}

	pageList := make([]int64, totalPages)
	var i int64
	for i = 0; i < totalPages; i++ {
		pageList[i] = i + 1
	}

	this.Data["keilVersions"] = versions
	this.Data["count"] = count
	this.Data["pages"] = pageNum
	this.Data["pageList"] = pageList
	this.Data["totalPages"] = totalPages
	this.Data["prevPage"] = pageNum - 1
	this.Data["nextPage"] = pageNum + 1
	this.Data["menu"] = "admin"
	this.Layout = "index.tpl"
	this.TplName = "admin/keil_versions.tpl"
}

// KeilVersionCreate POST /admin/keil-versions
func (this *AdminController) KeilVersionCreate() {
	versionName := this.GetString("versionName")
	uv4Path := this.GetString("uv4Path")
	if versionName == "" || uv4Path == "" {
		this.Data["json"] = map[string]interface{}{"code": 400, "message": "版本名称和路径不能为空"}
		this.ServeJSON()
		return
	}

	o := orm.NewOrm()
	kv := models.KeilVersion{
		VersionName: versionName,
		Uv4Path:     uv4Path,
	}
	id, err := o.Insert(&kv)
	if err != nil {
		this.Data["json"] = map[string]interface{}{"code": 500, "message": "添加失败：" + err.Error()}
		this.ServeJSON()
		return
	}
	this.Data["json"] = map[string]interface{}{"code": 0, "message": "ok", "data": map[string]interface{}{"id": id}}
	this.ServeJSON()
}

// KeilVersionUpdate PUT /admin/keil-versions/:id
func (this *AdminController) KeilVersionUpdate() {
	idStr := this.Ctx.Input.Param(":id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		this.Data["json"] = map[string]interface{}{"code": 400, "message": "无效的 ID"}
		this.ServeJSON()
		return
	}

	versionName := this.GetString("versionName")
	uv4Path := this.GetString("uv4Path")
	if versionName == "" || uv4Path == "" {
		this.Data["json"] = map[string]interface{}{"code": 400, "message": "版本名称和路径不能为空"}
		this.ServeJSON()
		return
	}

	o := orm.NewOrm()
	kv := models.KeilVersion{Id: id}
	if err := o.Read(&kv); err != nil {
		this.Data["json"] = map[string]interface{}{"code": 404, "message": "记录不存在"}
		this.ServeJSON()
		return
	}
	kv.VersionName = versionName
	kv.Uv4Path = uv4Path
	if _, err := o.Update(&kv, "VersionName", "Uv4Path"); err != nil {
		this.Data["json"] = map[string]interface{}{"code": 500, "message": "更新失败：" + err.Error()}
		this.ServeJSON()
		return
	}
	this.Data["json"] = map[string]interface{}{"code": 0, "message": "ok"}
	this.ServeJSON()
}

// KeilVersionDelete DELETE /admin/keil-versions/:id
func (this *AdminController) KeilVersionDelete() {
	idStr := this.Ctx.Input.Param(":id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		this.Data["json"] = map[string]interface{}{"code": 400, "message": "无效的 ID"}
		this.ServeJSON()
		return
	}

	o := orm.NewOrm()
	if o.QueryTable("repo_config").Filter("KeilVersionId", id).Exist() {
		this.Data["json"] = map[string]interface{}{"code": 400, "message": "该版本已被仓库引用，无法删除"}
		this.ServeJSON()
		return
	}

	if _, err := o.Delete(&models.KeilVersion{Id: id}); err != nil {
		this.Data["json"] = map[string]interface{}{"code": 500, "message": "删除失败：" + err.Error()}
		this.ServeJSON()
		return
	}
	this.Data["json"] = map[string]interface{}{"code": 0, "message": "ok"}
	this.ServeJSON()
}

// KeilVersionValidatePath POST /admin/keil-versions/validate-path
func (this *AdminController) KeilVersionValidatePath() {
	path := this.GetString("path")
	if path == "" {
		this.Data["json"] = map[string]interface{}{"code": 400, "message": "path 不能为空"}
		this.ServeJSON()
		return
	}
	_, err := os.Stat(path)
	exists := err == nil
	this.Data["json"] = map[string]interface{}{
		"code":    0,
		"message": "ok",
		"data":    map[string]interface{}{"exists": exists},
	}
	this.ServeJSON()
}

// Settings GET /admin/settings
func (this *AdminController) Settings() {
	o := orm.NewOrm()
	keys := []string{
		models.ConfigKeySMTPHost,
		models.ConfigKeySMTPPort,
		models.ConfigKeySMTPUser,
		models.ConfigKeySMTPPass,
		models.ConfigKeySMTPFrom,
		models.ConfigKeyAppBaseURL,
		models.ConfigKeyCommitMsgPattern,
		models.ConfigKeyAIProvider,
		models.ConfigKeyAIModel,
		models.ConfigKeyAIPrompt,
		models.ConfigKeyPermissionMode,
	}
	kv := make(map[string]string)
	for _, k := range keys {
		var c models.SysConfig
		if err := o.QueryTable("sys_config").Filter("ConfigKey", k).One(&c); err == nil {
			kv[k] = c.ConfigVal
		}
	}
	this.Data["cfg"] = kv
	this.Data["menu"] = "admin"
	this.Layout = "index.tpl"
	this.TplName = "admin/settings.tpl"
}

// SaveSettings POST /admin/settings
func (this *AdminController) SaveSettings() {
	o := orm.NewOrm()
	fields := map[string]bool{
		models.ConfigKeySMTPHost:           false,
		models.ConfigKeySMTPPort:           false,
		models.ConfigKeySMTPUser:           false,
		models.ConfigKeySMTPPass:           true,
		models.ConfigKeySMTPFrom:           false,
		models.ConfigKeyAppBaseURL:         false,
		models.ConfigKeyCommitMsgPattern:   false,
		models.ConfigKeyAIProvider:         false,
		models.ConfigKeyAIApiKey:           true,
		models.ConfigKeyAIModel:            false,
		models.ConfigKeyAIPrompt:           false,
		models.ConfigKeyPermissionMode:     false,
	}
	for k, isSecret := range fields {
		val := this.GetString(k)
		if val == "" {
			continue
		}
		// AI API Key 保存前加密
		if k == models.ConfigKeyAIApiKey {
			encrypted, err := aiPkg.Encrypt(val)
			if err == nil {
				val = encrypted
			}
		}
		var c models.SysConfig
		err := o.QueryTable("sys_config").Filter("ConfigKey", k).One(&c)
		if err != nil {
			c = models.SysConfig{ConfigKey: k, IsSecret: isSecret}
		}
		c.ConfigVal = val
		if c.Id == 0 {
			o.Insert(&c)
		} else {
			o.Update(&c, "ConfigVal", "UpdatedAt")
		}
	}
	this.Data["json"] = map[string]interface{}{"code": 0, "message": "ok"}
	this.ServeJSON()
}

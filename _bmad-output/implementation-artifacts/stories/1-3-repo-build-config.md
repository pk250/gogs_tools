---
story_id: 1-3
story_key: NOKEY-1-3
title: 仓库编译配置
epic: epic-1
status: ready-for-dev
created: '2026-03-25'
---

# Story 1-3：仓库编译配置

## 目标

为每个仓库提供编译参数配置页面，让开发者/项目负责人可以设置 Keil 版本、触发策略、产物文件名及通知方式，系统据此处理 Webhook 推送事件。

---

## 依赖（必须先完成）

| Story | 原因 |
|-------|------|
| Story 1-1 | `RepoConfig` 模型（`models/repo.go`）必须存在 |
| Story 1-2 | `KeilVersion` 数据已可读取，Keil版本下拉列表需要从该表查询 |

---

## 现有代码库上下文（必读）

### 仓库数据来源

**文件：** `models/gogs.go`（第 79-96 行）

`GogsDB` 表存储所有 Webhook 推送记录，`Repository_Name` 字段即仓库名。仓库列表页通过 `SELECT DISTINCT Repository_Name FROM gogs_d_b` 获取已知仓库。

```go
// 查询所有已知仓库名（去重）
var repoNames []orm.Params
o.Raw("SELECT DISTINCT repository_name FROM gogs_d_b ORDER BY repository_name").Values(&repoNames)
```

**注意 ORM 表名规则：** Beego ORM 自动将驼峰式 `GogsDB` → `gogs_d_b`，`RepoConfig` → `repo_config`，`KeilVersion` → `keil_version`。

### RepoConfig 模型（Story 1-1 创建）

**文件：** `models/repo.go`

```go
type RepoConfig struct {
    Id             int64
    RepoName       string    `orm:"size(128);unique"`
    KeilVersionId  int64     `orm:"default(0)"`        // 0 = 未配置
    TriggerMode    string    `orm:"size(16);default(manual)"`  // auto | manual
    ArtifactName   string    `orm:"size(128);null"`
    NotifyEmails   string    `orm:"type(text);null"`
    WebhookEnabled bool      `orm:"default(false)"`
    WebhookUrl     string    `orm:"size(512);null"`
    LintConfigPath string    `orm:"size(512);null"`
    CreatedAt      time.Time `orm:"auto_now_add;type(datetime)"`
    UpdatedAt      time.Time `orm:"auto_now;type(datetime)"`
}
```

### KeilVersion 模型（Story 1-2 创建）

**文件：** `models/keil_version.go`

```go
type KeilVersion struct {
    Id          int64
    VersionName string    `orm:"size(64);unique"`
    Uv4Path     string    `orm:"size(512)"`
    CreatedAt   time.Time `orm:"auto_now_add;type(datetime)"`
    UpdatedAt   time.Time `orm:"auto_now;type(datetime)"`
}
```

### 现有控制器模式

**文件：** `controllers/base.go`（第 10-38 行）

所有控制器必须嵌入 `BaseController`，其 `Prepare()` 已处理登录校验和用户角色注入：
- `this.Data["role"]` = "管理员" or "普通用户"
- `this.Data["username"]` = 用户名
- 未登录用户自动重定向至 `/login`

### 现有布局注册方式

**文件：** `controllers/layout.go`（参考实现）

```go
func (this *LayoutController) Datainfo() {
    this.Data["menu"] = "datainfo"
    this.Layout = "index.tpl"
    this.TplName = "datainfo.tpl"
}
```

### 路由注册方式

**文件：** `routers/router.go`

```go
beego.AutoRouter(&controllers.ListsController{})   // 自动路由模式
beego.Router("/login", &controllers.LoginController{})  // 手动路由模式
```

**本 Story 使用手动路由**（路径语义更清晰）：
```go
beego.Router("/repos", &controllers.RepoController{}, "get:List")
beego.Router("/repos/:repoName/config", &controllers.RepoController{}, "get:Config;post:SaveConfig")
```

### Toastr Toast 通知

`views/index.tpl` 已引入 Toastr.js，可直接使用：
```javascript
toastr.success('配置已保存', '成功', {timeOut: 3000});
toastr.error('保存失败', '错误');
```

---

## 验收标准

- [ ] 仓库列表页 `/repos` 显示所有已知仓库（来自 GogsDB 去重），每行展示：仓库名、Keil版本、触发策略、操作按钮
- [ ] 未配置的仓库显示「未配置」灰色标签
- [ ] 点击「配置」跳转至 `/repos/:repoName/config`
- [ ] 配置页显示表单：Keil版本下拉（从 keil_version 表加载）、触发策略（自动/手动）、产物文件名（默认仓库名）
- [ ] 配置页显示通知配置：邮件收件人（逗号分隔）、启用 Webhook 回调（checkbox）、Webhook URL
- [ ] 保存成功显示绿色 Toast 提示；保存失败显示红色 Toast
- [ ] 当前 MVP 阶段（宽松模式）：所有已登录用户可配置
- [ ] 配置保存使用 upsert 逻辑：存在则更新，不存在则插入

---

## 技术实现规范

### 文件结构

```
新建：
  controllers/repo.go           ← RepoController
  views/repo/list.tpl           ← 仓库列表页
  views/repo/config.tpl         ← 仓库配置页
修改：
  routers/router.go             ← 追加2条路由
  views/index.tpl               ← 追加「仓库配置」侧边栏菜单项
```

### `controllers/repo.go` 实现规范

```go
package controllers

import (
    "gogs_tools/models"

    "github.com/astaxie/beego/orm"
)

type RepoController struct {
    BaseController
}

// List GET /repos
// 展示仓库列表：从 GogsDB 取不重复仓库名，关联查询 RepoConfig 配置摘要
func (this *RepoController) List() {
    o := orm.NewOrm()

    // 1. 获取所有已知仓库名（来自 webhook 记录）
    var repoNames []orm.Params
    o.Raw("SELECT DISTINCT repository_name FROM gogs_d_b ORDER BY repository_name").Values(&repoNames)

    // 2. 查询所有已有配置
    var configs []models.RepoConfig
    o.QueryTable("repo_config").All(&configs)

    // 3. 构建 map 便于模板查找
    configMap := make(map[string]models.RepoConfig)
    for _, c := range configs {
        configMap[c.RepoName] = c
    }

    // 4. 查询 KeilVersion 名称 map（用于展示版本名而非 ID）
    var versions []models.KeilVersion
    o.QueryTable("keil_version").All(&versions)
    versionMap := make(map[int64]string)
    for _, v := range versions {
        versionMap[v.Id] = v.VersionName
    }

    this.Data["repoNames"] = repoNames
    this.Data["configMap"] = configMap
    this.Data["versionMap"] = versionMap
    this.Data["menu"] = "repos"
    this.Layout = "index.tpl"
    this.TplName = "repo/list.tpl"
}

// Config GET /repos/:repoName/config
// 展示单个仓库配置表单
func (this *RepoController) Config() {
    repoName := this.Ctx.Input.Param(":repoName")
    o := orm.NewOrm()

    // 加载现有配置（若不存在则使用默认值）
    config := models.RepoConfig{RepoName: repoName}
    err := o.Read(&config, "RepoName")
    if err == orm.ErrNoRows {
        config = models.RepoConfig{
            RepoName:    repoName,
            TriggerMode: "manual",
            ArtifactName: repoName,
        }
    }

    // 加载所有 Keil 版本（用于下拉选择）
    var versions []models.KeilVersion
    o.QueryTable("keil_version").All(&versions)

    this.Data["config"] = config
    this.Data["versions"] = versions
    this.Data["repoName"] = repoName
    this.Data["menu"] = "repos"
    this.Layout = "index.tpl"
    this.TplName = "repo/config.tpl"
}

// SaveConfig POST /repos/:repoName/config
// 保存仓库配置（upsert）
func (this *RepoController) SaveConfig() {
    repoName := this.Ctx.Input.Param(":repoName")
    o := orm.NewOrm()

    // 读取表单数据
    keilVersionId, _ := this.GetInt64("keil_version_id")
    triggerMode := this.GetString("trigger_mode")
    artifactName := this.GetString("artifact_name")
    notifyEmails := this.GetString("notify_emails")
    webhookEnabled, _ := this.GetBool("webhook_enabled")
    webhookUrl := this.GetString("webhook_url")
    lintConfigPath := this.GetString("lint_config_path")

    // 默认产物名为仓库名
    if artifactName == "" {
        artifactName = repoName
    }

    config := models.RepoConfig{
        RepoName:       repoName,
        KeilVersionId:  keilVersionId,
        TriggerMode:    triggerMode,
        ArtifactName:   artifactName,
        NotifyEmails:   notifyEmails,
        WebhookEnabled: webhookEnabled,
        WebhookUrl:     webhookUrl,
        LintConfigPath: lintConfigPath,
    }

    // Upsert：尝试读取已有记录
    exist := models.RepoConfig{RepoName: repoName}
    err := o.Read(&exist, "RepoName")
    if err == orm.ErrNoRows {
        _, err = o.Insert(&config)
    } else {
        config.Id = exist.Id
        _, err = o.Update(&config, "KeilVersionId", "TriggerMode", "ArtifactName",
            "NotifyEmails", "WebhookEnabled", "WebhookUrl", "LintConfigPath")
    }

    if err != nil {
        this.Data["json"] = map[string]interface{}{"code": 500, "message": "保存失败: " + err.Error()}
    } else {
        this.Data["json"] = map[string]interface{}{"code": 0, "message": "保存成功"}
    }
    this.ServeJSON()
}
```

### `routers/router.go` 修改规范

在 `init()` 函数末尾追加：

```go
beego.Router("/repos", &controllers.RepoController{}, "get:List")
beego.Router("/repos/:repoName/config", &controllers.RepoController{}, "get:Config;post:SaveConfig")
```

**注意：** `:repoName` 是路径参数，在 controller 中用 `this.Ctx.Input.Param(":repoName")` 读取（不是 `GetString`）。

### `views/index.tpl` 侧边栏追加

在现有侧边栏菜单中追加「仓库配置」菜单项（参考已有菜单项的 HTML 结构）：

```html
<li class="{{if eq .menu "repos"}}active{{end}}">
    <a href="/repos"><i class="fa fa-code-fork fa-fw"></i> <span class="nav-label">仓库配置</span></a>
</li>
```

---

## `views/repo/list.tpl` 实现规范

**目录注意：** 必须先在 `views/` 下创建 `repo/` 子目录，再创建模板文件。

```html
<!-- 继承主布局：this.Layout = "index.tpl" -->
<div class="row">
  <div class="col-lg-12">
    <div class="ibox">
      <div class="ibox-title">
        <h5>仓库编译配置</h5>
      </div>
      <div class="ibox-content">
        <table class="table table-striped table-hover">
          <thead>
            <tr>
              <th>仓库名称</th>
              <th>Keil 版本</th>
              <th>触发策略</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            {{range .repoNames}}
            {{$name := index . "repository_name"}}
            <tr>
              <td>{{$name}}</td>
              <td>
                {{$cfg := index $.configMap $name}}
                {{if $cfg.KeilVersionId}}
                  {{index $.versionMap $cfg.KeilVersionId}}
                {{else}}
                  <span class="label label-default">未配置</span>
                {{end}}
              </td>
              <td>
                {{if $cfg.TriggerMode}}
                  {{if eq $cfg.TriggerMode "auto"}}
                    <span class="label label-primary">自动</span>
                  {{else}}
                    <span class="label label-default">手动</span>
                  {{end}}
                {{else}}
                  <span class="label label-default">未配置</span>
                {{end}}
              </td>
              <td>
                <a href="/repos/{{$name}}/config" class="btn btn-xs btn-white">配置</a>
              </td>
            </tr>
            {{end}}
          </tbody>
        </table>
      </div>
    </div>
  </div>
</div>
```

---

## `views/repo/config.tpl` 实现规范

```html
<!-- 继承主布局 -->
<div class="row">
  <div class="col-lg-8">
    <div class="ibox">
      <div class="ibox-title">
        <h5>仓库配置 - {{.repoName}}</h5>
      </div>
      <div class="ibox-content">
        <form id="configForm">
          <div class="form-group">
            <label>Keil 版本</label>
            <select name="keil_version_id" class="form-control">
              <option value="0">-- 请选择 Keil 版本 --</option>
              {{range .versions}}
              <option value="{{.Id}}" {{if eq $.config.KeilVersionId .Id}}selected{{end}}>{{.VersionName}}</option>
              {{end}}
            </select>
          </div>
          <div class="form-group">
            <label>触发策略</label>
            <select name="trigger_mode" class="form-control">
              <option value="manual" {{if eq .config.TriggerMode "manual"}}selected{{end}}>手动触发</option>
              <option value="auto" {{if eq .config.TriggerMode "auto"}}selected{{end}}>自动触发（Webhook 推送时自动编译）</option>
            </select>
          </div>
          <div class="form-group">
            <label>产物文件名</label>
            <input type="text" name="artifact_name" class="form-control" value="{{.config.ArtifactName}}" placeholder="{{.repoName}}">
            <p class="help-block">编译产物的基础文件名（不含扩展名），默认为仓库名</p>
          </div>
          <hr>
          <h4>通知配置</h4>
          <div class="form-group">
            <label>邮件收件人</label>
            <input type="text" name="notify_emails" class="form-control" value="{{.config.NotifyEmails}}" placeholder="a@example.com,b@example.com">
            <p class="help-block">多个邮箱用英文逗号分隔</p>
          </div>
          <div class="form-group">
            <div class="checkbox">
              <label>
                <input type="checkbox" name="webhook_enabled" value="true" {{if .config.WebhookEnabled}}checked{{end}}>
                启用 Webhook 回调
              </label>
            </div>
          </div>
          <div class="form-group" id="webhookUrlGroup" style="{{if not .config.WebhookEnabled}}display:none{{end}}">
            <label>Webhook URL</label>
            <input type="text" name="webhook_url" class="form-control" value="{{.config.WebhookUrl}}" placeholder="https://your-server/callback">
          </div>
          <div class="form-group">
            <button type="button" class="btn btn-primary" onclick="saveConfig()">保存配置</button>
            <a href="/repos" class="btn btn-default">返回列表</a>
          </div>
        </form>
      </div>
    </div>
  </div>
</div>

<script>
// checkbox 联动显示/隐藏 Webhook URL 输入框
$("input[name='webhook_enabled']").change(function() {
    if ($(this).is(':checked')) {
        $('#webhookUrlGroup').show();
    } else {
        $('#webhookUrlGroup').hide();
    }
});

function saveConfig() {
    var formData = {
        keil_version_id: $("select[name='keil_version_id']").val(),
        trigger_mode: $("select[name='trigger_mode']").val(),
        artifact_name: $("input[name='artifact_name']").val(),
        notify_emails: $("input[name='notify_emails']").val(),
        webhook_enabled: $("input[name='webhook_enabled']").is(':checked'),
        webhook_url: $("input[name='webhook_url']").val()
    };

    $.post('/repos/{{.repoName}}/config', formData, function(resp) {
        if (resp.code === 0) {
            toastr.success('配置已保存', '成功', {timeOut: 3000});
        } else {
            toastr.error(resp.message, '保存失败');
        }
    }).fail(function() {
        toastr.error('请求失败，请重试', '错误');
    });
}
</script>
```

---

## 关键陷阱与注意事项

### 1. ORM 表名与字段名转换

Beego ORM 自动将 Go 驼峰转蛇形：
- `GogsDB` → 表名 `gogs_d_b`（特殊！非 `gogs_db`）
- `RepoConfig` → 表名 `repo_config`
- `KeilVersionId` → 字段 `keil_version_id`

原始 SQL 查询必须使用实际表名：
```go
o.Raw("SELECT DISTINCT repository_name FROM gogs_d_b").Values(&repoNames)
```

### 2. 路径参数读取方式

```go
// 正确：Beego 路由参数
repoName := this.Ctx.Input.Param(":repoName")

// 错误：这只读取 Query String
repoName := this.GetString("repoName")
```

### 3. Upsert 逻辑顺序

`models.RepoConfig` 的 `RepoName` 字段有 `unique` 约束，直接 Insert 重复记录会报错。必须先 Read，再决定 Insert 或 Update：
```go
exist := models.RepoConfig{RepoName: repoName}
err := o.Read(&exist, "RepoName")  // 按 RepoName 字段查询（非主键）
if err == orm.ErrNoRows {
    _, err = o.Insert(&config)       // 新增
} else {
    config.Id = exist.Id
    _, err = o.Update(&config, ...)  // 更新指定字段
}
```

### 4. Go 模板中的 map 索引

Beego 模板（基于 html/template）中，访问 map 用 `index`：
```html
{{$cfg := index $.configMap $name}}   <!-- 正确 -->
{{$.configMap[$name]}}                 <!-- 语法错误 -->
```

### 5. `views/repo/` 子目录

Beego 默认从 `views/` 目录查找模板。模板路径 `repo/config.tpl` 表示 `views/repo/config.tpl`，需要手动创建 `views/repo/` 目录。

### 6. POST 表单 vs JSON 请求体

`SaveConfig` 使用标准 HTML Form POST（非 JSON body），所以用 `this.GetString()`、`this.GetInt64()`、`this.GetBool()` 读取，而非 `json.Unmarshal`。

---

## 实现顺序建议

1. 确认 Story 1-1 已执行（`repo_config`、`keil_version` 表已存在）
2. 创建 `controllers/repo.go`（先实现 `List`，再实现 `Config`、`SaveConfig`）
3. 在 `routers/router.go` 追加两条路由
4. 创建 `views/repo/` 目录
5. 创建 `views/repo/list.tpl`
6. 创建 `views/repo/config.tpl`
7. 在 `views/index.tpl` 追加侧边栏菜单项
8. 编译运行，访问 `/repos` 验证列表页
9. 点击「配置」验证表单加载已有数据
10. 提交表单验证 Toast 提示和 upsert 逻辑

---

## 未实现（留待后续 Story）

| 功能 | 对应 Story |
|------|------------|
| 宽松/严格权限模式切换 | Story 3-1 |
| Webhook URL 回调通知 | Story 3-3 |
| PC-Lint 配置文件路径管理 | Story 3-4 |
| `LintConfigPath` 字段在本 Story 中渲染表单但不验证文件存在性 | Story 2-1 |

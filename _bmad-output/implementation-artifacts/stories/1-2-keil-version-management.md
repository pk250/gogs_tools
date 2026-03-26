---
story_id: 1-2
story_key: NOKEY-1-2
title: Keil 版本管理（管理员）
epic: epic-1
status: ready-for-dev
created: '2026-03-25'
depends_on:
  - 1-1
---

# Story 1-2：Keil 版本管理（管理员）

## 目标

为管理员提供 Keil 版本注册与管理功能，支持添加、编辑、删除、路径验证和分页列表，为后续仓库配置（Story 1-3）绑定 Keil 版本提供数据支撑。

---

## 前置依赖

- **Story 1-1 必须先完成**：`models/keil_version.go` 中的 `KeilVersion` 模型已创建并注册到 ORM。

---

## 现有代码库上下文（必读）

### 现有文件结构

```
controllers/
├── base.go        ← BaseController，含 Prepare() 鉴权逻辑
├── layout.go      ← LayoutController，含页面渲染示例
├── login.go
├── logout.go
├── register.go
├── index.go
├── lists.go
├── gogs.go
└── api.go

views/
├── index.tpl      ← 主布局（所有页面的 Layout）
├── datainfo.tpl   ← 现有页面示例（分页实现参考）
├── compile.tpl
├── knowledge.tpl
├── login.tpl
└── register.tpl

routers/router.go  ← 路由注册入口
models/users.go    ← init() 注册所有模型
```

### BaseController 鉴权机制（必须继承）

**文件：** `controllers/base.go`（第 10-38 行）

```go
func (this *BaseController) Prepare() {
    users := this.Ctx.Input.Session("UserData")
    if users == nil &&
        this.Ctx.Request.RequestURI != "/login" &&
        this.Ctx.Request.RequestURI != "/register" &&
        this.Ctx.Request.RequestURI != "/gogs" {
        this.Ctx.Redirect(302, "/login")
    } else if users != nil {
        o := orm.NewOrm()
        user := models.Users{Id: users.(models.Users).Id}
        err := o.Read(&user)
        if err != nil {
            this.Ctx.Redirect(302, "/login")
        } else {
            if user.IsAdmin {
                this.Data["role"] = "管理员"
            } else {
                this.Data["role"] = "普通用户"
            }
            this.Data["Title"] = beego.AppConfig.String("Title")
            this.Data["username"] = user.Username
        }
    }
}
```

**关键规则：**
- `AdminController` 必须嵌入 `BaseController`（不是直接用 `beego.Controller`）
- 管理员权限检查在 `AdminController.Prepare()` 中叠加判断 `IsAdmin`，不满足则 302 到首页
- `this.Data["role"]`, `this.Data["username"]`, `this.Data["Title"]` 由 `BaseController.Prepare()` 自动注入，无需重复设置

### 分页实现参考

**文件：** `controllers/layout.go`（第 23-56 行，`Datainfo()` 方法）

```go
pages := this.Ctx.Input.Params()  // URL path 参数
qs := o.QueryTable("datainfos")
count, _ := qs.Count()
this.Data["count"] = count
if len(pages) > 0 {
    num, _ := strconv.ParseInt(pages["0"], 10, 64)
    qs.Limit(10, 10*(num-1)).All(&datainfos)
    this.Data["pages"] = num
} else {
    qs.Limit(10, 0).All(&datainfos)
    this.Data["pages"] = 1
}
```

### 路由注册方式（参考）

**文件：** `routers/router.go`

```go
beego.Router("/", &controllers.IndexController{})
beego.AutoRouter(&controllers.LayoutController{})  // 自动路由：/layout/datainfo 等
```

**本 Story 使用显式路由**（Admin 需要精确路径控制，不用 AutoRouter）：

```go
// 新增路由示例（不要用 AutoRouter for admin）
beego.Router("/admin/keil-versions", &controllers.AdminController{}, "get:KeilVersionList")
```

### 布局模板使用方式

**文件：** `controllers/layout.go`（第 54-55 行）

```go
this.Layout = "index.tpl"       // 布局文件（侧边栏、导航）
this.TplName = "datainfo.tpl"   // 内容区模板
```

`views/index.tpl` 中侧边栏菜单用 `{{if eq "xxx" .menu}}active{{end}}` 控制高亮，本 Story 使用 `menu = "admin"`。

---

## 验收标准

- [ ] 管理员访问 `/admin/keil-versions` 可看到 Keil 版本列表（分页，每页 10 条）
- [ ] 列表显示：版本名称、UV4.exe 路径、创建时间、操作（编辑/删除）
- [ ] 管理员可添加 Keil 版本（版本号 + UV4.exe 完整路径）
- [ ] 添加/编辑时点击「验证路径」按钮，服务端检查文件是否存在，内联显示 ✅/❌ 反馈
- [ ] 可编辑版本名称和路径
- [ ] 删除时若有 `repo_configs` 记录引用该版本，返回错误提示，不执行删除
- [ ] 删除时若无引用，正常删除并刷新列表
- [ ] 非管理员访问 `/admin/*` 任何路由均 302 跳转首页
- [ ] 所有 API 响应使用统一 JSON 格式：`{"code": 0, "message": "ok", "data": ...}`

---

## 技术实现规范

### 新增文件

```
controllers/admin.go             ← 新建：AdminController
views/admin/keil_versions.tpl   ← 新建：Keil版本管理页面
```

### 修改文件

```
routers/router.go               ← 追加 admin 路由
views/index.tpl                 ← 追加侧边栏「系统管理」菜单项（仅 admin 可见）
```

---

### `controllers/admin.go` 实现规范

```go
package controllers

import (
    "gogs_tools/models"
    "os"
    "strconv"

    "github.com/astaxie/beego"
    "github.com/astaxie/beego/orm"
)

type AdminController struct {
    BaseController
}

// Prepare 叠加管理员权限校验
func (this *AdminController) Prepare() {
    this.BaseController.Prepare()  // 先执行登录校验
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

// KeilVersionList GET /admin/keil-versions/:page
func (this *AdminController) KeilVersionList() {
    // 分页逻辑（参考 layout.go Datainfo()）
    // this.Data["menu"] = "admin"
    // this.Layout = "index.tpl"
    // this.TplName = "admin/keil_versions.tpl"
}

// KeilVersionCreate POST /admin/keil-versions
// 返回 JSON: {"code":0,"message":"ok","data":{"id":1}}
func (this *AdminController) KeilVersionCreate() {
    // 读取 versionName, uv4Path 参数
    // 校验非空
    // INSERT KeilVersion
    // 返回 JSON
}

// KeilVersionUpdate PUT /admin/keil-versions/:id
func (this *AdminController) KeilVersionUpdate() {
    // 读取 id（path参数），versionName, uv4Path（body）
    // UPDATE KeilVersion
    // 返回 JSON
}

// KeilVersionDelete DELETE /admin/keil-versions/:id
// 删除前检查 RepoConfig 引用
func (this *AdminController) KeilVersionDelete() {
    // 读取 id
    // 检查：o.QueryTable("repo_config").Filter("KeilVersionId", id).Exist()
    // 若存在引用：返回 {"code":400,"message":"该版本已被仓库引用，无法删除"}
    // 否则：DELETE，返回 {"code":0,"message":"ok"}
}

// KeilVersionValidatePath POST /admin/keil-versions/validate-path
// 请求体: {"path": "/path/to/UV4.exe"}
// 响应: {"code":0,"data":{"exists":true}}
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
```

**重要注意事项：**

1. `Prepare()` 中先调 `this.BaseController.Prepare()`，再做 Admin 检查，避免未登录用户绕过
2. `os.Stat(path)` 检查文件存在性，这是服务端本地路径（UV4.exe 在服务器上），不是 HTTP 请求
3. Beego ORM 表名由模型名自动推导：`KeilVersion` → `keil_version`，`RepoConfig` → `repo_config`
4. 分页参数通过 path param 传递（`:page`），与现有 `Datainfo()` 保持一致

---

### 路由注册（`routers/router.go` 追加）

```go
// Admin 路由 —— 精确路由，不使用 AutoRouter
beego.Router("/admin/keil-versions", &controllers.AdminController{}, "get:KeilVersionList")
beego.Router("/admin/keil-versions/:page", &controllers.AdminController{}, "get:KeilVersionList")
beego.Router("/admin/keil-versions", &controllers.AdminController{}, "post:KeilVersionCreate")
beego.Router("/admin/keil-versions/:id", &controllers.AdminController{}, "put:KeilVersionUpdate;delete:KeilVersionDelete")
beego.Router("/admin/keil-versions/validate-path", &controllers.AdminController{}, "post:KeilVersionValidatePath")
```

**注意：** `/validate-path` 路由必须注册在 `/:id` 之前，否则 Beego 会把 `validate-path` 匹配为 `:id`。

---

### `views/admin/keil_versions.tpl` 实现规范

```html
<!-- 继承布局：this.Layout = "index.tpl"，内容区由此文件填充 -->
<div class="row">
  <div class="col-lg-12">
    <div class="ibox">
      <div class="ibox-title">
        <h5>Keil 版本管理</h5>
        <button class="btn btn-primary btn-sm pull-right" onclick="showAddModal()">+ 添加版本</button>
      </div>
      <div class="ibox-content">
        <!-- 版本列表表格 -->
        <table class="table table-striped">
          <thead>
            <tr>
              <th>版本名称</th>
              <th>UV4.exe 路径</th>
              <th>创建时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            {{range .keilVersions}}
            <tr>
              <td>{{.VersionName}}</td>
              <td>{{.Uv4Path}}</td>
              <td>{{.CreatedAt.Format "2006-01-02 15:04"}}</td>
              <td>
                <button class="btn btn-xs btn-default" onclick="showEditModal({{.Id}}, '{{.VersionName}}', '{{.Uv4Path}}')">编辑</button>
                <button class="btn btn-xs btn-danger" onclick="deleteVersion({{.Id}})">删除</button>
              </td>
            </tr>
            {{end}}
          </tbody>
        </table>

        <!-- 分页控件（参考 datainfo.tpl 分页实现） -->
        <nav>
          <ul class="pagination">
            {{range $i := pages .count 10}}
            <li class="{{if eq $i $.pages}}active{{end}}">
              <a href="/admin/keil-versions/{{$i}}">{{$i}}</a>
            </li>
            {{end}}
          </ul>
        </nav>
      </div>
    </div>
  </div>
</div>

<!-- 添加/编辑 Modal -->
<div class="modal fade" id="keilModal">
  <div class="modal-dialog">
    <div class="modal-content">
      <div class="modal-header">
        <h4 class="modal-title" id="modalTitle">添加 Keil 版本</h4>
      </div>
      <div class="modal-body">
        <input type="hidden" id="keilId" value="">
        <div class="form-group">
          <label>版本名称</label>
          <input type="text" class="form-control" id="versionName" placeholder="例：Keil 5.38">
        </div>
        <div class="form-group">
          <label>UV4.exe 路径</label>
          <div class="input-group">
            <input type="text" class="form-control" id="uv4Path" placeholder="例：C:\Keil_v5\UV4\UV4.exe">
            <span class="input-group-btn">
              <button class="btn btn-default" onclick="validatePath()">验证路径</button>
            </span>
          </div>
          <span id="pathValidateResult" class="help-block"></span>
        </div>
      </div>
      <div class="modal-footer">
        <button class="btn btn-default" data-dismiss="modal">取消</button>
        <button class="btn btn-primary" onclick="saveKeilVersion()">保存</button>
      </div>
    </div>
  </div>
</div>

<script>
function showAddModal() {
    $('#keilId').val('');
    $('#versionName').val('');
    $('#uv4Path').val('');
    $('#pathValidateResult').text('');
    $('#modalTitle').text('添加 Keil 版本');
    $('#keilModal').modal('show');
}

function showEditModal(id, name, path) {
    $('#keilId').val(id);
    $('#versionName').val(name);
    $('#uv4Path').val(path);
    $('#pathValidateResult').text('');
    $('#modalTitle').text('编辑 Keil 版本');
    $('#keilModal').modal('show');
}

function validatePath() {
    var path = $('#uv4Path').val();
    $.post('/admin/keil-versions/validate-path', {path: path}, function(res) {
        if (res.code === 0) {
            var ok = res.data.exists;
            $('#pathValidateResult').html(ok ? '<span class="text-success">✅ 路径有效</span>' : '<span class="text-danger">❌ 文件不存在</span>');
        }
    });
}

function saveKeilVersion() {
    var id = $('#keilId').val();
    var versionName = $('#versionName').val();
    var uv4Path = $('#uv4Path').val();
    if (!versionName || !uv4Path) {
        toastr.error('版本名称和路径不能为空');
        return;
    }
    if (id) {
        // 编辑
        $.ajax({
            url: '/admin/keil-versions/' + id,
            type: 'PUT',
            data: {versionName: versionName, uv4Path: uv4Path},
            success: function(res) {
                if (res.code === 0) {
                    toastr.success('保存成功');
                    $('#keilModal').modal('hide');
                    location.reload();
                } else {
                    toastr.error(res.message);
                }
            }
        });
    } else {
        // 新增
        $.post('/admin/keil-versions', {versionName: versionName, uv4Path: uv4Path}, function(res) {
            if (res.code === 0) {
                toastr.success('添加成功');
                $('#keilModal').modal('hide');
                location.reload();
            } else {
                toastr.error(res.message);
            }
        });
    }
}

function deleteVersion(id) {
    if (!confirm('确认删除此 Keil 版本？')) return;
    $.ajax({
        url: '/admin/keil-versions/' + id,
        type: 'DELETE',
        success: function(res) {
            if (res.code === 0) {
                toastr.success('删除成功');
                location.reload();
            } else {
                toastr.error(res.message);
            }
        }
    });
}
</script>
```

**模板注意事项：**
- 复用 `views/index.tpl` 中已有的 Toastr、Bootstrap、jQuery（无需额外引入）
- 分页 helper 函数需在 Beego 模板函数中注册，或直接在控制器中计算页码列表传入模板
- `{{.CreatedAt.Format "2006-01-02 15:04"}}` 是 Go time 格式，Beego 模板直接支持

---

### `views/index.tpl` 侧边栏修改

在侧边栏 `<ul class="nav metismenu">` 中追加管理员菜单项（仅管理员可见）：

```html
{{if eq .role "管理员"}}
<li class="{{if eq .menu "admin"}}active{{end}}">
    <a href="#"><i class="fa fa-cogs fa-fw"></i> <span class="nav-label">系统管理</span> <span class="fa arrow"></span></a>
    <ul class="nav nav-second-level">
        <li><a href="/admin/keil-versions">Keil 版本管理</a></li>
    </ul>
</li>
{{end}}
```

---

## 关键约束与陷阱

1. **Beego 路由顺序**：`/admin/keil-versions/validate-path` 必须在 `/admin/keil-versions/:id` 之前注册，否则 `validate-path` 会被当作 `:id`。

2. **HTTP 方法覆盖**：浏览器 `<form>` 只支持 GET/POST，jQuery `$.ajax` 可发 PUT/DELETE。确认前端用 `$.ajax({type:'PUT'})` 和 `$.ajax({type:'DELETE'})`。

3. **ORM 表名**：Beego ORM 自动将 `KeilVersion` 转为表名 `keil_version`（下划线小写）。`QueryTable()` 传字符串时用 `"keil_version"`，或传 `new(KeilVersion)` 更安全。

4. **路径分隔符**：UV4.exe 路径在 Windows 上用反斜杠，`os.Stat()` 在 Windows 可直接处理，无需转换。

5. **`views/admin/` 目录**：需要手动创建 `views/admin/` 目录，再放置 `keil_versions.tpl`。

6. **Beego 模板分页**：Beego 不内置分页 range helper，推荐在控制器计算页码切片 `[]int64` 传入模板，而非在模板中用函数。

---

## 本地测试要点

1. 以管理员账号登录（`IsAdmin=true` 的用户）
2. 访问 `/admin/keil-versions`，确认显示列表页
3. 用普通用户账号访问 `/admin/keil-versions`，确认 302 跳回首页
4. 添加一条记录（路径可填系统中存在的任意 `.exe` 文件用于测试）
5. 点击「验证路径」，填真实路径显示 ✅，填不存在路径显示 ❌
6. 编辑修改版本名称，保存后刷新确认更新
7. 删除无引用记录，确认成功；创建一条 `repo_config` 引用该版本后再删除，确认返回错误提示

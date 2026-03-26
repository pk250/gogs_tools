<div class="wrapper wrapper-content">
    <div class="row">
        <div class="col-lg-12">
            <div class="ibox">
                <div class="ibox-title">
                    <h5>用户管理</h5>
                    <div class="ibox-tools">
                        <span class="text-muted">共 {{.count}} 个用户</span>
                    </div>
                </div>
                <div class="ibox-content">
                    <table class="table table-striped">
                        <thead>
                            <tr>
                                <th>ID</th>
                                <th>用户名</th>
                                <th>邮箱</th>
                                <th>注册时间</th>
                                <th>最后登录</th>
                                <th>登录次数</th>
                                <th>管理员</th>
                                <th>状态</th>
                                <th>操作</th>
                            </tr>
                        </thead>
                        <tbody>
                            {{range .users}}
                            <tr id="user-row-{{.Id}}">
                                <td>{{.Id}}</td>
                                <td>{{.Username}}</td>
                                <td>{{.Email}}</td>
                                <td>{{.Created.Format "2006-01-02 15:04"}}</td>
                                <td>{{.LastLogin.Format "2006-01-02 15:04"}}</td>
                                <td>{{.LoginCount}}</td>
                                <td>
                                    <span id="admin-badge-{{.Id}}" class="label {{if .IsAdmin}}label-primary{{else}}label-default{{end}}">
                                        {{if .IsAdmin}}管理员{{else}}普通用户{{end}}
                                    </span>
                                </td>
                                <td>
                                    <span id="active-badge-{{.Id}}" class="label {{if .IsActive}}label-success{{else}}label-danger{{end}}">
                                        {{if .IsActive}}正常{{else}}已禁用{{end}}
                                    </span>
                                </td>
                                <td>
                                    <button class="btn btn-xs btn-default" onclick="toggleAdmin({{.Id}})">
                                        <span id="admin-btn-{{.Id}}">{{if .IsAdmin}}取消管理员{{else}}设为管理员{{end}}</span>
                                    </button>
                                    <button class="btn btn-xs {{if .IsActive}}btn-warning{{else}}btn-success{{end}}" onclick="toggleActive({{.Id}})">
                                        <span id="active-btn-{{.Id}}">{{if .IsActive}}禁用{{else}}启用{{end}}</span>
                                    </button>
                                </td>
                            </tr>
                            {{end}}
                        </tbody>
                    </table>

                    {{if gt .totalPages 1}}
                    <nav>
                        <ul class="pagination">
                            {{if gt .page 1}}
                            <li class="page-item">
                                <a class="page-link" href="/admin/users?page={{.prevPage}}">&laquo;</a>
                            </li>
                            {{end}}
                            {{range .pageList}}
                            <li class="page-item {{if eq . $.page}}active{{end}}">
                                <a class="page-link" href="/admin/users?page={{.}}">{{.}}</a>
                            </li>
                            {{end}}
                            {{if lt .page .totalPages}}
                            <li class="page-item">
                                <a class="page-link" href="/admin/users?page={{.nextPage}}">&raquo;</a>
                            </li>
                            {{end}}
                        </ul>
                    </nav>
                    {{end}}
                </div>
            </div>
        </div>
    </div>
</div>

<script>
function toggleAdmin(uid) {
    $.post('/admin/users/' + uid + '/toggle-admin', {}, function(res) {
        if (res.code === 0) {
            var badge = document.getElementById('admin-badge-' + uid);
            var btn = document.getElementById('admin-btn-' + uid);
            if (res.isAdmin) {
                badge.className = 'label label-primary';
                badge.textContent = '管理员';
                btn.textContent = '取消管理员';
            } else {
                badge.className = 'label label-default';
                badge.textContent = '普通用户';
                btn.textContent = '设为管理员';
            }
            toastr.success('权限已更新');
        } else {
            toastr.error(res.message);
        }
    });
}

function toggleActive(uid) {
    $.post('/admin/users/' + uid + '/toggle-active', {}, function(res) {
        if (res.code === 0) {
            var badge = document.getElementById('active-badge-' + uid);
            var btn = document.querySelector('#user-row-' + uid + ' button:last-child');
            if (res.isActive) {
                badge.className = 'label label-success';
                badge.textContent = '正常';
                btn.className = 'btn btn-xs btn-warning';
                document.getElementById('active-btn-' + uid).textContent = '禁用';
            } else {
                badge.className = 'label label-danger';
                badge.textContent = '已禁用';
                btn.className = 'btn btn-xs btn-success';
                document.getElementById('active-btn-' + uid).textContent = '启用';
            }
            toastr.success('状态已更新');
        } else {
            toastr.error(res.message);
        }
    });
}
</script>

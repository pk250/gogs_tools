<div class="wrapper wrapper-content">
    <div class="row">
        <div class="col-lg-12">
            <div class="ibox">
                <div class="ibox-title">
                    <h5>Keil 版本管理</h5>
                    <div class="ibox-tools">
                        <button class="btn btn-primary btn-sm" onclick="showAddModal()">+ 添加版本</button>
                    </div>
                </div>
                <div class="ibox-content">
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

                    <nav>
                        <ul class="pagination">
                            {{if gt .pages 1}}
                            <li class="page-item">
                                <a class="page-link" href="/admin/keil-versions/{{.prevPage}}">&laquo;</a>
                            </li>
                            {{end}}
                            {{range .pageList}}
                            <li class="page-item {{if eq . $.pages}}active{{end}}">
                                <a class="page-link" href="/admin/keil-versions/{{.}}">{{.}}</a>
                            </li>
                            {{end}}
                            {{if lt .pages .totalPages}}
                            <li class="page-item">
                                <a class="page-link" href="/admin/keil-versions/{{.nextPage}}">&raquo;</a>
                            </li>
                            {{end}}
                        </ul>
                    </nav>
                </div>
            </div>
        </div>
    </div>
</div>

<!-- 添加/编辑模态框 -->
<div class="modal fade" id="keilModal" tabindex="-1" role="dialog">
    <div class="modal-dialog" role="document">
        <div class="modal-content">
            <div class="modal-header">
                <h5 class="modal-title" id="modalTitle">添加 Keil 版本</h5>
                <button type="button" class="close" data-dismiss="modal">&times;</button>
            </div>
            <div class="modal-body">
                <input type="hidden" id="keilId">
                <div class="form-group">
                    <label>版本名称</label>
                    <input type="text" class="form-control" id="versionName" placeholder="例：MDK5.38">
                </div>
                <div class="form-group">
                    <label>UV4.exe 路径</label>
                    <div class="input-group">
                        <input type="text" class="form-control" id="uv4Path" placeholder="例：C:\Keil_v5\UV4\UV4.exe">
                        <span class="input-group-btn">
                            <button class="btn btn-default" type="button" onclick="validatePath()">验证路径</button>
                        </span>
                    </div>
                    <small id="pathValidateResult" class="form-text"></small>
                </div>
            </div>
            <div class="modal-footer">
                <button type="button" class="btn btn-default" data-dismiss="modal">取消</button>
                <button type="button" class="btn btn-primary" onclick="saveKeilVersion()">保存</button>
            </div>
        </div>
    </div>
</div>

<div class="footer">
    <div>
        <strong>Copyright</strong> Dakewe &copy; 2023-2033
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
            $('#pathValidateResult').html(ok ? '<span class="text-success">路径有效</span>' : '<span class="text-danger">文件不存在</span>');
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

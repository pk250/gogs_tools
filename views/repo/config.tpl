<div class="wrapper wrapper-content">
<div class="row">
  <div class="col-lg-8">
    <div class="ibox">
      <div class="ibox-title">
        <h5>仓库配置 - {{.repoName}}</h5>
      </div>
      <div class="ibox-content">
        {{if not .canEdit}}
        <div class="alert alert-warning"><i class="fa fa-lock"></i> 当前为<strong>严格模式</strong>，仅管理员或项目负责人可修改仓库配置。</div>
        {{end}}
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
          <div class="form-group" id="webhookUrlGroup" {{if not .config.WebhookEnabled}}style="display:none"{{end}}>
            <label>Webhook URL</label>
            <input type="text" name="webhook_url" class="form-control" value="{{.config.WebhookUrl}}" placeholder="https://your-server/callback">
          </div>
          <div class="form-group">
            <button type="button" class="btn btn-primary" onclick="saveConfig()" {{if not .canEdit}}disabled{{end}}>保存配置</button>
            <a href="/repos" class="btn btn-default">返回列表</a>
          </div>
        </form>

        <hr>
        <h4>PC-Lint 配置文件</h4>
        {{if .lintFileName}}
        <div class="alert alert-info">
          <i class="fa fa-file"></i> 当前文件：<strong>{{.lintFileName}}</strong>
          {{if .lintUploadedAt}}&nbsp;<small class="text-muted">上传时间：{{.lintUploadedAt.Format "2006-01-02 15:04"}}</small>{{end}}
          {{if .canEdit}}
          <button class="btn btn-xs btn-danger m-l-sm" onclick="deleteLintConfig()">删除</button>
          {{end}}
        </div>
        {{else}}
        <p class="text-muted">尚未上传 .lnt 配置文件，编译后将跳过 PC-Lint 检查。</p>
        {{end}}
        {{if .lintTplURL}}
        <p><a href="{{.lintTplURL}}" class="btn btn-xs btn-default"><i class="fa fa-download"></i> 下载系统默认模板</a></p>
        {{end}}
        {{if .canEdit}}
        <form id="lintUploadForm" enctype="multipart/form-data">
          <div class="input-group" style="max-width:400px;">
            <input type="file" name="lint_file" accept=".lnt" class="form-control" id="lintFileInput">
            <span class="input-group-btn">
              <button type="button" class="btn btn-primary" onclick="uploadLintConfig()">上传</button>
            </span>
          </div>
          <p class="help-block">仅支持 .lnt 文件，大小 ≤ 1MB</p>
        </form>
        {{end}}
      </div>
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
$('input[name="webhook_enabled"]').change(function() {
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

function uploadLintConfig() {
    var fileInput = document.getElementById('lintFileInput');
    if (!fileInput || !fileInput.files.length) {
        toastr.warning('请选择 .lnt 文件', '提示');
        return;
    }
    var fd = new FormData();
    fd.append('lint_file', fileInput.files[0]);
    fetch('/repos/{{.repoName}}/lint-config', {method: 'POST', body: fd})
        .then(function(r){ return r.json(); })
        .then(function(d){
            if (d.code === 0) {
                toastr.success('上传成功', '成功', {timeOut: 2000});
                setTimeout(function(){ location.reload(); }, 1500);
            } else {
                toastr.error(d.message, '上传失败');
            }
        }).catch(function(){ toastr.error('请求失败', '错误'); });
}

function deleteLintConfig() {
    if (!confirm('确定删除已上传的 .lnt 配置文件？删除后编译将跳过 PC-Lint 检查。')) return;
    fetch('/repos/{{.repoName}}/lint-config', {method: 'DELETE'})
        .then(function(r){ return r.json(); })
        .then(function(d){
            if (d.code === 0) {
                toastr.success('已删除', '成功', {timeOut: 2000});
                setTimeout(function(){ location.reload(); }, 1500);
            } else {
                toastr.error(d.message, '删除失败');
            }
        }).catch(function(){ toastr.error('请求失败', '错误'); });
}
</script>

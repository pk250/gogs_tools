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
</script>

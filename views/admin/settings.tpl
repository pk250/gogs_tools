<div class="wrapper wrapper-content">
<div class="row">
  <div class="col-lg-8 col-lg-offset-2">
    <div class="ibox">
      <div class="ibox-title"><h5>系统设置 &mdash; 邮件通知</h5></div>
      <div class="ibox-content">
        <form id="settings-form">
          <div class="form-group">
            <label>SMTP 主机</label>
            <input type="text" class="form-control" name="smtp_host" value="{{index .cfg "smtp_host"}}" placeholder="smtp.example.com">
          </div>
          <div class="form-group">
            <label>SMTP 端口</label>
            <input type="text" class="form-control" name="smtp_port" value="{{index .cfg "smtp_port"}}" placeholder="465">
          </div>
          <div class="form-group">
            <label>SMTP 用户名</label>
            <input type="text" class="form-control" name="smtp_user" value="{{index .cfg "smtp_user"}}" placeholder="user@example.com">
          </div>
          <div class="form-group">
            <label>SMTP 密码</label>
            <input type="password" class="form-control" name="smtp_pass" placeholder="（留空则不修改）">
          </div>
          <div class="form-group">
            <label>发件人地址</label>
            <input type="text" class="form-control" name="smtp_from" value="{{index .cfg "smtp_from"}}" placeholder="noreply@example.com">
          </div>
          <div class="form-group">
            <label>应用基础 URL（用于邮件中的链接）</label>
            <input type="text" class="form-control" name="app_base_url" value="{{index .cfg "app_base_url"}}" placeholder="http://yourserver:8080">
          </div>
          <button type="submit" class="btn btn-primary">保存</button>
          <span id="save-msg" style="margin-left:12px;display:none;"></span>
        </form>
      </div>
    </div>

    <div class="ibox">
      <div class="ibox-title"><h5>代码质量 &mdash; Git 提交规范</h5></div>
      <div class="ibox-content">
        <form id="git-check-form">
          <div class="form-group">
            <label>Commit Message 正则规范</label>
            <input type="text" class="form-control" name="commit_msg_pattern" value="{{index .cfg "commit_msg_pattern"}}" placeholder="例：^(feat|fix|docs|chore)(\(.+\))?: .+">
            <p class="help-block">留空则跳过 commit message 规范检查。使用 Go 正则语法。</p>
          </div>
          <button type="submit" class="btn btn-primary">保存</button>
          <span id="git-save-msg" style="margin-left:12px;display:none;"></span>
        </form>
      </div>
    </div>
  </div>
</div>
</div>
<div class="footer">
    <div><strong>Copyright</strong> Dakewe &copy; 2023-2033</div>
</div>
<script>
document.getElementById('settings-form').addEventListener('submit', function(e) {
    e.preventDefault();
    var data = new URLSearchParams(new FormData(this));
    fetch('/admin/settings', {method:'POST', body: data})
        .then(function(r){ return r.json(); })
        .then(function(res) {
            var msg = document.getElementById('save-msg');
            msg.style.display = 'inline';
            if (res.code === 0) {
                msg.className = 'text-success';
                msg.textContent = '保存成功';
            } else {
                msg.className = 'text-danger';
                msg.textContent = res.message;
            }
            setTimeout(function(){ msg.style.display='none'; }, 3000);
        });
});

document.getElementById('git-check-form').addEventListener('submit', function(e) {
    e.preventDefault();
    var data = new URLSearchParams(new FormData(this));
    fetch('/admin/settings', {method:'POST', body: data})
        .then(function(r){ return r.json(); })
        .then(function(res) {
            var msg = document.getElementById('git-save-msg');
            msg.style.display = 'inline';
            if (res.code === 0) {
                msg.className = 'text-success';
                msg.textContent = '保存成功';
            } else {
                msg.className = 'text-danger';
                msg.textContent = res.message;
            }
            setTimeout(function(){ msg.style.display='none'; }, 3000);
        });
});
</script>

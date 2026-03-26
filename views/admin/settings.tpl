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
      <div class="ibox-title"><h5>代码质量 &mdash; AI 代码审查</h5></div>
      <div class="ibox-content">
        <form id="ai-form">
          <div class="form-group">
            <label>AI 服务商</label>
            <select class="form-control" name="ai_provider">
              <option value="" {{if eq (index .cfg "ai_provider") ""}}selected{{end}}>-- 不启用 --</option>
              <option value="claude" {{if eq (index .cfg "ai_provider") "claude"}}selected{{end}}>Anthropic Claude</option>
              <option value="openai" {{if eq (index .cfg "ai_provider") "openai"}}selected{{end}}>OpenAI</option>
            </select>
          </div>
          <div class="form-group">
            <label>API Key（留空则不修改）</label>
            <input type="password" class="form-control" name="ai_api_key" placeholder="****">
          </div>
          <div class="form-group">
            <label>模型名称</label>
            <input type="text" class="form-control" name="ai_model" value="{{index .cfg "ai_model"}}" placeholder="claude-sonnet-4-6 / gpt-4o">
          </div>
          <div class="form-group">
            <label>审查提示词</label>
            <textarea class="form-control" name="ai_prompt" rows="3" placeholder="请对以下代码变更进行审查...">{{index .cfg "ai_prompt"}}</textarea>
          </div>
          <div class="form-group">
            <label>API Base URL（可选，兼容 OpenAI 接口的自定义地址）</label>
            <input type="text" class="form-control" name="ai_base_url" value="{{index .cfg "ai_base_url"}}" placeholder="https://api.openai.com/v1">
            <p class="help-block">留空则使用服务商默认地址。</p>
          </div>
          <button type="submit" class="btn btn-primary">保存</button>
          <span id="ai-save-msg" style="margin-left:12px;display:none;"></span>
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

    <div class="ibox">
      <div class="ibox-title"><h5>系统管理 &mdash; 权限模式</h5></div>
      <div class="ibox-content">
        <form id="permission-form">
          <div class="form-group">
            <label>仓库配置权限模式</label>
            <select class="form-control" name="permission_mode">
              <option value="loose" {{if ne (index .cfg "permission_mode") "strict"}}selected{{end}}>宽松模式（所有用户可配置仓库）</option>
              <option value="strict" {{if eq (index .cfg "permission_mode") "strict"}}selected{{end}}>严格模式（仅管理员/项目负责人可配置）</option>
            </select>
            <p class="help-block">切换后即时生效，无需重启。</p>
          </div>
          <button type="submit" class="btn btn-primary">保存</button>
          <span id="perm-save-msg" style="margin-left:12px;display:none;"></span>
        </form>
      </div>
    </div>

    <div class="ibox">
      <div class="ibox-title"><h5>代码质量 &mdash; PC-Lint</h5></div>
      <div class="ibox-content">
        <form id="pclint-form">
          <div class="form-group">
            <label>PC-Lint 可执行文件路径</label>
            <input type="text" class="form-control" name="pclint_exe" value="{{index .cfg "pclint_exe"}}" placeholder="/usr/bin/lint-nt">
            <p class="help-block">留空则不启用 PC-Lint 静态检查。</p>
          </div>
          <button type="submit" class="btn btn-primary">保存</button>
          <span id="pclint-save-msg" style="margin-left:12px;display:none;"></span>
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

document.getElementById('ai-form').addEventListener('submit', function(e) {
    e.preventDefault();
    var data = new URLSearchParams(new FormData(this));
    fetch('/admin/settings', {method:'POST', body: data})
        .then(function(r){ return r.json(); })
        .then(function(res) {
            var msg = document.getElementById('ai-save-msg');
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

document.getElementById('permission-form').addEventListener('submit', function(e) {
    e.preventDefault();
    var data = new URLSearchParams(new FormData(this));
    fetch('/admin/settings', {method:'POST', body: data})
        .then(function(r){ return r.json(); })
        .then(function(res) {
            var msg = document.getElementById('perm-save-msg');
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

document.getElementById('pclint-form').addEventListener('submit', function(e) {
    e.preventDefault();
    var data = new URLSearchParams(new FormData(this));
    fetch('/admin/settings', {method:'POST', body: data})
        .then(function(r){ return r.json(); })
        .then(function(res) {
            var msg = document.getElementById('pclint-save-msg');
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

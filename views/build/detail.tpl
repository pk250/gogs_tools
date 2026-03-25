<div class="wrapper wrapper-content">
<div class="row">
  <div class="col-lg-12">
    <div id="ws-alert" class="alert alert-warning" style="display:none;">
      连接已断开，正在重连...
    </div>

    <div class="ibox">
      <div class="ibox-title">
        <h5>任务 #{{.task.Id}} &mdash; {{.task.RepoName}}</h5>
        <span id="task-status" class="label label-{{.statusClass}}">{{.task.Status}}</span>
        {{if and (eq .triggerMode "manual") (eq .task.Status "pending")}}
        <button id="btn-trigger" class="btn btn-xs btn-primary m-l-sm" onclick="triggerBuild()">触发编译</button>
        {{end}}
        {{if and .hasWebhook (or (eq .task.Status "success") (eq .task.Status "failed"))}}
        <button class="btn btn-xs btn-default m-l-sm" onclick="retryWebhook()">重试 Webhook</button>
        {{end}}
      </div>
      <div class="ibox-content">
        <p>Commit: <code>{{if ge (len .task.CommitHash) 7}}{{slice .task.CommitHash 0 7}}{{else}}{{.task.CommitHash}}{{end}}</code> &nbsp; 提交人: {{.task.Author}}</p>
        <p>{{.task.CommitMsg}}</p>

        <!-- Phase progress bar -->
        <div class="row m-t-sm">
          <div class="col-lg-12">
            <div class="progress">
              {{if eq .task.Status "pending"}}
              <div class="progress-bar" style="width:5%">待队列</div>
              {{else if eq .task.Status "running"}}
              <div class="progress-bar progress-bar-warning progress-bar-striped active" style="width:33%">编译中</div>
              {{else if eq .task.Status "failed"}}
              <div class="progress-bar progress-bar-danger" style="width:33%">编译失败</div>
              {{else if eq .task.Status "success"}}
              <div class="progress-bar progress-bar-success" style="width:33%">编译</div>
              <div class="progress-bar progress-bar-success" style="width:33%">质检</div>
              <div class="progress-bar progress-bar-success" style="width:34%">通知</div>
              {{end}}
            </div>
            <small class="text-muted">
              阶段：
              {{$done := or (eq .task.Status "success") (eq .task.Status "failed")}}
              <span class="{{if or (eq .task.Status "running") $done}}text-success{{end}}"><i class="fa fa-cog"></i> 编译</span>
              &rarr;
              <span class="{{if and $done .hasLint}}{{if eq .lintResult.Status "fail"}}text-danger{{else}}text-success{{end}}{{end}}"><i class="fa fa-search"></i> PC-Lint</span>
              &rarr;
              <span class="{{if and $done .hasAI}}{{if eq .aiResult.Status "fail"}}text-danger{{else}}text-success{{end}}{{end}}"><i class="fa fa-magic"></i> AI审查</span>
              &rarr;
              <span class="{{if and $done .hasGitCheck}}{{if eq .gitCheckResult.Status "fail"}}text-danger{{else}}text-success{{end}}{{end}}"><i class="fa fa-check-circle"></i> Git规范</span>
              &rarr;
              <span class="{{if eq .task.Status "success"}}text-success{{end}}"><i class="fa fa-envelope"></i> 通知</span>
            </small>
          </div>
        </div>
      </div>
    </div>

    <div class="ibox">
      <div class="ibox-title"><h5>编译日志</h5></div>
      <div class="ibox-content">
        <pre id="log-box" style="height:500px;overflow-y:auto;background:#1a1a2e;color:#e0e0e0;padding:12px;"></pre>
      </div>
    </div>

    {{if .artifacts}}
    <div class="ibox">
      <div class="ibox-title"><h5>编译产物</h5></div>
      <div class="ibox-content">
        <ul class="list-group">
          {{range .artifacts}}
          <li class="list-group-item">
            <a href="/build/artifacts/{{$.task.Id}}/{{.}}"><i class="fa fa-download"></i> {{.}}</a>
          </li>
          {{end}}
        </ul>
      </div>
    </div>
    {{end}}

    {{if .hasLint}}
    <div class="ibox">
      <div class="ibox-title" style="cursor:pointer;" onclick="togglePanel('lint-body', this)">
        <h5>PC-Lint 检查结果
          {{if eq .lintResult.Status "pass"}}<span class="label label-success">通过</span>{{end}}
          {{if eq .lintResult.Status "warn"}}<span class="label label-warning">警告</span>{{end}}
          {{if eq .lintResult.Status "fail"}}<span class="label label-danger">失败</span>{{end}}
          {{if eq .lintResult.Status "skip"}}<span class="label label-default">跳过</span>{{end}}
        </h5>
        <div class="ibox-tools"><i class="fa fa-chevron-{{if or (eq .lintResult.Status "fail") (eq .lintResult.Status "warn")}}up{{else}}down{{end}}" id="lint-icon"></i></div>
      </div>
      <div class="ibox-content" id="lint-body" style="{{if and (ne .lintResult.Status "fail") (ne .lintResult.Status "warn")}}display:none;{{end}}">
        <p>{{.lintResult.Summary}}</p>
        {{if .lintResult.Detail}}
        <pre style="max-height:300px;overflow-y:auto;background:#1a1a2e;color:#e0e0e0;padding:12px;">{{.lintResult.Detail}}</pre>
        {{end}}
      </div>
    </div>
    {{end}}

    {{if .hasGitCheck}}
    <div class="ibox">
      <div class="ibox-title" style="cursor:pointer;" onclick="togglePanel('git-body', this)">
        <h5>Git 提交规范
          {{if eq .gitCheckResult.Status "pass"}}<span class="label label-success">合规</span>{{end}}
          {{if eq .gitCheckResult.Status "fail"}}<span class="label label-danger">不合规</span>{{end}}
          {{if eq .gitCheckResult.Status "skip"}}<span class="label label-default">跳过</span>{{end}}
        </h5>
        <div class="ibox-tools"><i class="fa fa-chevron-{{if eq .gitCheckResult.Status "fail"}}up{{else}}down{{end}}" id="git-icon"></i></div>
      </div>
      <div class="ibox-content" id="git-body" style="{{if ne .gitCheckResult.Status "fail"}}display:none;{{end}}">
        <p>{{.gitCheckResult.Summary}}</p>
      </div>
    </div>
    {{end}}

    {{if .hasAI}}
    <div class="ibox">
      <div class="ibox-title" style="cursor:pointer;" onclick="togglePanel('ai-body', this)">
        <h5>AI 代码审查
          {{if eq .aiResult.Status "pass"}}<span class="label label-success">完成</span>{{end}}
          {{if eq .aiResult.Status "fail"}}<span class="label label-danger">失败</span>{{end}}
          {{if eq .aiResult.Status "skip"}}<span class="label label-default">跳过</span>{{end}}
        </h5>
        <div class="ibox-tools"><i class="fa fa-chevron-{{if eq .aiResult.Status "fail"}}up{{else}}down{{end}}" id="ai-icon"></i></div>
      </div>
      <div class="ibox-content" id="ai-body" style="{{if ne .aiResult.Status "fail"}}display:none;{{end}}">
        <p>{{.aiResult.Summary}}</p>
        {{if .aiResult.Detail}}
        <pre style="max-height:400px;overflow-y:auto;background:#f8f9fa;color:#333;padding:12px;white-space:pre-wrap;">{{.aiResult.Detail}}</pre>
        {{end}}
      </div>
    </div>
    {{end}}

  </div>
</div>
</div>
<div class="footer">
    <div>
        <strong>Copyright</strong> Dakewe &copy; 2023-2033
    </div>
</div>
<script>
function togglePanel(bodyId, titleEl) {
    var body = document.getElementById(bodyId);
    var icon = titleEl.querySelector('i[id$="-icon"]');
    if (body.style.display === 'none') {
        body.style.display = '';
        if (icon) { icon.className = icon.className.replace('fa-chevron-down', 'fa-chevron-up'); }
    } else {
        body.style.display = 'none';
        if (icon) { icon.className = icon.className.replace('fa-chevron-up', 'fa-chevron-down'); }
    }
}
</script>
<script>
(function() {
    var taskId = {{.task.Id}};
    var logBox = document.getElementById('log-box');
    var wsAlert = document.getElementById('ws-alert');
    var ws;
    var autoScroll = true;
    var reconnectTimer;

    logBox.addEventListener('scroll', function() {
        var atBottom = logBox.scrollHeight - logBox.scrollTop <= logBox.clientHeight + 20;
        autoScroll = atBottom;
    });

    function appendLine(text) {
        logBox.textContent += text + '\n';
        if (autoScroll) {
            logBox.scrollTop = logBox.scrollHeight;
        }
    }

    function connect() {
        var proto = location.protocol === 'https:' ? 'wss' : 'ws';
        ws = new WebSocket(proto + '://' + location.host + '/ws/build/' + taskId);

        ws.onopen = function() {
            wsAlert.style.display = 'none';
            clearTimeout(reconnectTimer);
        };

        ws.onmessage = function(e) {
            var msg = JSON.parse(e.data);
            if (msg.type === 'log') {
                appendLine(msg.data);
            } else if (msg.type === 'status') {
                var el = document.getElementById('task-status');
                if (el) el.textContent = msg.data;
            } else if (msg.type === 'complete') {
                appendLine('[编译完成：' + msg.data + ']');
                autoScroll = false;
                ws.close();
                setTimeout(function(){ location.reload(); }, 1500);
            } else if (msg.type === 'error') {
                appendLine('[错误：' + msg.data + ']');
            }
        };

        ws.onerror = function() {
            wsAlert.style.display = 'block';
        };

        ws.onclose = function() {
            wsAlert.style.display = 'block';
            reconnectTimer = setTimeout(connect, 3000);
        };
    }

    connect();
})();

function triggerBuild() {
    var btn = document.getElementById('btn-trigger');
    if (!btn) return;
    btn.disabled = true;
    btn.textContent = '触发中...';
    fetch('/api/build/{{.task.Id}}/enqueue', {method: 'POST'})
        .then(function(r){ return r.json(); })
        .then(function(d){
            if (d.code === 0) {
                btn.textContent = '已触发';
                setTimeout(function(){ location.reload(); }, 800);
            } else {
                btn.textContent = '触发失败';
                btn.disabled = false;
                alert(d.message);
            }
        })
        .catch(function(){
            btn.textContent = '触发失败';
            btn.disabled = false;
        });
}

function retryWebhook() {
    fetch('/api/build/{{.task.Id}}/webhook-retry', {method: 'POST'})
        .then(function(r){ return r.json(); })
        .then(function(d){
            alert(d.code === 0 ? 'Webhook 回调已触发' : d.message);
        })
        .catch(function(){ alert('请求失败'); });
}
</script>

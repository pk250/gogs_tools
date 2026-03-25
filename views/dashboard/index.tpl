<div class="wrapper wrapper-content">
<div class="row">
  <!-- Stats cards -->
  <div class="col-lg-4">
    <div class="ibox">
      <div class="ibox-content">
        <h1 class="no-margins text-success">{{.todaySuccess}}</h1>
        <small>今日编译成功</small>
      </div>
    </div>
  </div>
  <div class="col-lg-4">
    <div class="ibox">
      <div class="ibox-content">
        <h1 class="no-margins text-danger">{{.todayFailed}}</h1>
        <small>今日编译失败</small>
      </div>
    </div>
  </div>
  <div class="col-lg-4">
    <div class="ibox">
      <div class="ibox-content">
        <h1 class="no-margins text-warning">{{.todayRunning}}</h1>
        <small>编译进行中</small>
      </div>
    </div>
  </div>
</div>

<div class="row">
  <div class="col-lg-12">
    <div class="ibox">
      <div class="ibox-title">
        <h5>编译任务列表</h5>
      </div>
      <div class="ibox-content">
        <!-- Filters -->
        <form method="get" action="/dashboard" class="form-inline m-b-sm">
          <div class="form-group m-r-sm">
            <select name="repo" class="form-control input-sm">
              <option value="">所有仓库</option>
              {{range .repoList}}
              <option value="{{.}}" {{if eq . $.filterRepo}}selected{{end}}>{{.}}</option>
              {{end}}
            </select>
          </div>
          <div class="form-group m-r-sm">
            <select name="status" class="form-control input-sm">
              <option value="">所有状态</option>
              <option value="pending" {{if eq .filterStatus "pending"}}selected{{end}}>待队列</option>
              <option value="running" {{if eq .filterStatus "running"}}selected{{end}}>编译中</option>
              <option value="success" {{if eq .filterStatus "success"}}selected{{end}}>成功</option>
              <option value="failed" {{if eq .filterStatus "failed"}}selected{{end}}>失败</option>
            </select>
          </div>
          <div class="form-group m-r-sm">
            <input type="text" name="author" class="form-control input-sm" placeholder="提交人" value="{{.filterAuthor}}">
          </div>
          <button type="submit" class="btn btn-sm btn-primary">筛选</button>
          <a href="/dashboard" class="btn btn-sm btn-white m-l-xs">重置</a>
        </form>

        <table class="table table-hover table-striped" id="task-table">
          <thead>
            <tr>
              <th>#</th>
              <th>仓库</th>
              <th>提交人</th>
              <th>Commit</th>
              <th>状态</th>
              <th>审查状态</th>
              <th>时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            {{range .tasks}}
            <tr id="row-{{.Id}}" data-status="{{.Status}}">
              <td>{{.Id}}</td>
              <td>{{.RepoName}}</td>
              <td>{{.Author}}</td>
              <td>
                <code title="{{.CommitHash}}">{{if ge (len .CommitHash) 7}}{{slice .CommitHash 0 7}}{{else}}{{.CommitHash}}{{end}}</code>
                <br><small class="text-muted">{{.CommitMsg}}</small>
              </td>
              <td>
                {{if eq .Status "pending"}}
                  <span class="label label-default">待队列</span>
                {{else if eq .Status "running"}}
                  <span class="label label-warning"><i class="fa fa-spinner fa-spin"></i> 编译中</span>
                {{else if eq .Status "success"}}
                  <span class="label label-success">成功</span>
                {{else if eq .Status "failed"}}
                  <span class="label label-danger">失败</span>
                {{end}}
              </td>
              <td>
                {{$rs := index $.taskReviewStatus .Id}}
                {{if $rs.Label}}
                  <span class="label label-{{$rs.Class}}">{{$rs.Label}}</span>
                {{else}}
                  <span class="label label-default">-</span>
                {{end}}
              </td>
              <td>
                <span title="{{.CreatedAt}}">{{.CreatedAt.Format "2006-01-02 15:04"}}</span>
              </td>
              <td>
                <a href="/build/detail/{{.Id}}" class="btn btn-xs btn-white">详情</a>
              </td>
            </tr>
            {{end}}
          </tbody>
        </table>

        <!-- Pagination -->
        <nav>
          <ul class="pagination pagination-sm">
            {{if gt .page 1}}
            <li><a href="/dashboard?page={{.prevPage}}&repo={{.filterRepo}}&status={{.filterStatus}}&author={{.filterAuthor}}">«</a></li>
            {{else}}
            <li class="disabled"><a>«</a></li>
            {{end}}
            <li class="disabled"><a>第 {{.page}} / {{.totalPages}} 页（共 {{.count}} 条）</a></li>
            {{if lt .page .totalPages}}
            <li><a href="/dashboard?page={{.nextPage}}&repo={{.filterRepo}}&status={{.filterStatus}}&author={{.filterAuthor}}">»</a></li>
            {{else}}
            <li class="disabled"><a>»</a></li>
            {{end}}
          </ul>
        </nav>
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
(function(){
    // Auto-refresh running rows via polling every 5s
    function refreshRunning() {
        var rows = document.querySelectorAll('tr[data-status="running"]');
        if (rows.length === 0) return;
        setTimeout(function(){ location.reload(); }, 5000);
    }
    refreshRunning();
})();
</script>

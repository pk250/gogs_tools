<div class="wrapper wrapper-content">
<div class="row">
  <div class="col-lg-12">
    <div class="ibox">
      <div class="ibox-title">
        <h5>团队视图 &mdash; 仓库配置总览</h5>
      </div>
      <div class="ibox-content">
        <!-- Member filter -->
        <form method="get" action="/admin/team" class="form-inline m-b-sm">
          <div class="form-group m-r-sm">
            <select name="member" class="form-control input-sm">
              <option value="">所有成员</option>
              {{range .memberList}}
              <option value="{{.}}" {{if eq . $.filterMember}}selected{{end}}>{{.}}</option>
              {{end}}
            </select>
          </div>
          <button type="submit" class="btn btn-sm btn-primary">筛选</button>
          <a href="/admin/team" class="btn btn-sm btn-white m-l-xs">重置</a>
        </form>

        <table class="table table-hover table-striped">
          <thead>
            <tr>
              <th>仓库名</th>
              <th>Keil 版本</th>
              <th>触发策略</th>
              <th>产物文件名</th>
              <th>成员（推送者）</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            {{range .repoList}}
            <tr>
              <td><strong>{{.RepoName}}</strong></td>
              <td>
                {{if .KeilName}}{{.KeilName}}{{else}}<span class="text-muted">未配置</span>{{end}}
              </td>
              <td>
                {{if not .HasConfig}}
                  <span class="text-muted">未配置</span>
                {{else if eq .Config.TriggerMode "auto"}}
                  <span class="label label-primary">自动</span>
                {{else}}
                  <span class="label label-default">手动</span>
                {{end}}
              </td>
              <td>
                {{if .Config.ArtifactName}}{{.Config.ArtifactName}}{{else}}<span class="text-muted">-</span>{{end}}
              </td>
              <td>
                {{range .Pushers}}
                  <span class="label label-info m-r-xs">{{.}}</span>
                {{end}}
                {{if not .Pushers}}<span class="text-muted">-</span>{{end}}
              </td>
              <td>
                <a href="/repos/{{.RepoName}}/config" class="btn btn-xs btn-white">编辑配置</a>
              </td>
            </tr>
            {{end}}
            {{if not .repoList}}
            <tr><td colspan="6" class="text-center text-muted">暂无数据</td></tr>
            {{end}}
          </tbody>
        </table>
      </div>
    </div>
  </div>
</div>
</div>
<div class="footer">
    <div><strong>Copyright</strong> Dakewe &copy; 2023-2033</div>
</div>

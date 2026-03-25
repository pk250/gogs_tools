<div class="wrapper wrapper-content">
<div class="row">
  <div class="col-lg-12">
    <div class="ibox">
      <div class="ibox-title">
        <h5>仓库编译配置</h5>
      </div>
      <div class="ibox-content">
        <table class="table table-striped table-hover">
          <thead>
            <tr>
              <th>仓库名称</th>
              <th>Keil 版本</th>
              <th>触发策略</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            {{range .repoNames}}
            {{$name := index . "repository_name"}}
            <tr>
              <td>{{$name}}</td>
              <td>
                {{$cfg := index $.configMap $name}}
                {{if $cfg.KeilVersionId}}
                  {{index $.versionMap $cfg.KeilVersionId}}
                {{else}}
                  <span class="label label-default">未配置</span>
                {{end}}
              </td>
              <td>
                {{if $cfg.TriggerMode}}
                  {{if eq $cfg.TriggerMode "auto"}}
                    <span class="label label-primary">自动</span>
                  {{else}}
                    <span class="label label-default">手动</span>
                  {{end}}
                {{else}}
                  <span class="label label-default">未配置</span>
                {{end}}
              </td>
              <td>
                <a href="/repos/{{$name}}/config" class="btn btn-xs btn-white">配置</a>
              </td>
            </tr>
            {{end}}
          </tbody>
        </table>
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

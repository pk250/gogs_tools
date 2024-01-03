<div class="wrapper wrapper-content">
    <div class="row">
        <div class="col-lg-12">
            <div class="ibox float-e-margins">
                <div class="ibox-title">
                    <h5>编译列表</h5>
                    <div class="ibox-tools">
                        <a class="collapse-link"><i class="fa fa-chevron-up"></i></a>
                        <a class="close-link"><i class="fa fa-times"></i></a>
                    </div>
                </div>
                <div class="ibox-content">
                    <div class="bootstrap-table bootstrap4">
                        <div class="fixed-table-toolbar">
                            <div class="columns columns-right btn-group float-right">
                                <button class="btn btn-secondary" type="button" name="refresh" aria-label="Refresh" title="刷新">
                                    <i class="fa fa-refresh"></i>
                                </button>
                            </div>
                            <div class="float-right search btn-group">
                                <input class="form-control search-input" type="text" placeholder="搜索" autocomplete="off">
                            </div>
                        </div>
                        <div class="fixed-table-container">
                            <table class="table table-striped">
                                <thead>
                                    <tr>
                                        <th class="footable-visible footable-sortable" style="text-align: center; ">
                                            仓库名称
                                        </th>
                                        <th class="footable-visible footable-sortable" style="text-align: center; ">
                                            commit
                                        </th>
                                        <th class="footable-visible footable-sortable" style="text-align: center; ">
                                            提交日期
                                        </th>
                                        <th class="footable-visible footable-sortable" style="text-align: center; ">
                                            提交作者
                                        </th>
                                        <th class="footable-visible footable-sortable" style="text-align: center; ">
                                            编译状态
                                        </th>
                                        <th class="footable-visible footable-sortable" style="text-align: center; ">
                                            编译账号
                                        </th>
                                        <th class="footable-visible footable-sortable" style="text-align: center; ">
                                            编译时间
                                        </th>
                                        <th class="footable-visible footable-sortable" style="text-align: center; ">
                                            操作
                                        </th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {{range .datainfo}}
                                    <tr class="footable-even">
                                        <td class="footable-visible" style="text-align: center; ">{{.StorageName}}</td>
                                        <td class="footable-visible" style="text-align: center; ">{{.CommitValue}}</td>
                                        <td class="footable-visible" style="text-align: center; ">{{.CommitTime}}</td>
                                        <td class="footable-visible" style="text-align: center; ">{{.CommitAuth}}</td>
                                        {{if .CompileStatus}}
                                        <td class="footable-visible" style="text-align: center; "><span class="label-primary">已编译</span></td>
                                        <td class="footable-visible" style="text-align: center; "><span class="label">{{.CompileUser}}</span></td>
                                        <td class="footable-visible" style="text-align: center; "><span class="label">{{.CompileTime}}</span></td>
                                        {{else}}
                                        <td class="footable-visible" style="text-align: center; "><span class="label">未编译</span></td>
                                        <td class="footable-visible" style="text-align: center; "><span class="label">Null</span></td>
                                        <td class="footable-visible" style="text-align: center; "><span class="label">Null</span></td>
                                        {{end}}
                                        <td class="footable-visible" style="text-align: center; ">
                                            <div class="btn-group">
                                                {{if .CompileStatus}}
                                                <a href="{{.CompilePath}}" class="btn btn-outline btn-primary">
                                                    下载
                                                </a>
                                                {{end}}
                                                <a href="#" class="btn btn-outline btn-primary">
                                                    编译
                                                </a>
                                            </div>
                                        </td>
                                    </tr>
                                    {{end}}
                                </tbody>
                            </table>
                            <div class="fixed-table-footer">
                                <table>
                                    <thead>
                                        <tr></tr>
                                    </thead>
                                </table>
                            </div>
                        </div>
                        <div class="fixed-table-pagination">
                            <div class="float-left pagination-detail"></div>
                            <div class="float-right pagination">
                                <button type="button" class="btn btn-white">
                                    <i class="fa fa-chevron-left"></i>
                                </button>
                                <button class="btn btn-primary active">1</button>
                                <button class="btn btn-white">2</button>
                                <button class="btn btn-white">3</button>
                                <button class="btn btn-white">4</button>
                                <button class="btn btn-white">5</button>
                                <button type="button" class="btn btn-white">
                                    <i class="fa fa-chevron-right"></i>
                                </button>
                            </div>
                        </div>
                    </div>
                    <div class="clearfix">

                    </div>
                </div>
            </div>
        </div>
    </div>
</div>
<script>

</script>
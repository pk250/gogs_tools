<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <title>{{ .Title }}</title>
    <link href="/static/css/bootstrap.min.css" rel="stylesheet">
    <link href="/static/font-awesome/css/font-awesome.css" rel="stylesheet">

    <!-- Toastr style -->
    <link href="/static/css/plugins/toastr/toastr.min.css" rel="stylesheet">

    <!-- Gritter -->
    <link href="/static/js/plugins/gritter/jquery.gritter.css" rel="stylesheet">

    <!-- FooTable -->
    <link href="/static/css/plugins/footable/footable.core.css" rel="stylesheet">

    <link href="/static/css/animate.css" rel="stylesheet">
    <link href="/static/css/style.css" rel="stylesheet">

    <!-- Mainly scripts -->
    <!-- <script src="/static/js/jquery-3.1.1.min.js"></script> -->
    <script src="https://cdn.bootcss.com/jquery/3.4.1/jquery.min.js"></script>
    <script src="/static/js/popper.min.js"></script>
    <script src="/static/js/bootstrap.js"></script>
    <script src="/static/js/plugins/metisMenu/jquery.metisMenu.js"></script>
    <script src="/static/js/plugins/slimscroll/jquery.slimscroll.min.js"></script>

    <!-- Flot -->
    <script src="/static/js/plugins/flot/jquery.flot.js"></script>
    <script src="/static/js/plugins/flot/jquery.flot.tooltip.min.js"></script>
    <script src="/static/js/plugins/flot/jquery.flot.spline.js"></script>
    <script src="/static/js/plugins/flot/jquery.flot.resize.js"></script>
    <script src="/static/js/plugins/flot/jquery.flot.pie.js"></script>

    <!-- Peity -->
    <script src="/static/js/plugins/peity/jquery.peity.min.js"></script>

    <!-- Custom and plugin javascript -->
    <script src="/static/js/inspinia.js"></script>
    <script src="/static/js/plugins/pace/pace.min.js"></script>

    <!-- jQuery UI -->
    <script src="/static/js/plugins/jquery-ui/jquery-ui.min.js"></script>

    <!-- GITTER -->
    <script src="/static/js/plugins/gritter/jquery.gritter.min.js"></script>

    <!-- Sparkline -->
    <script src="/static/js/plugins/sparkline/jquery.sparkline.min.js"></script>

    <!-- ChartJS-->
    <script src="/static/js/plugins/chartJs/Chart.min.js"></script>

    <!-- Toastr -->
    <script src="/static/js/plugins/toastr/toastr.min.js"></script>

     <!-- FooTable -->
     <script src="/static/js/plugins/footable/footable.all.min.js"></script>

</head>
<body>
    <div id="wrapper">
        <nav class="navbar-default navbar-static-side" role="navigation">
            <div class="sidebar-collapse">
                <ul class="nav metismenu" id="side-menu">
                    <li class="nav-header">
                        <div class="dropdown profile-element">
                            <!-- <img alt="image" class="rounded-circle" src="/static/img/profile_small.jpg"/> -->
                            <i class="rounded-circle fa fa-user fa-3x"></i>
                            <a data-toggle="dropdown" class="dropdown-toggle" href="#">
                                <span class="block m-t-xs font-bold">{{.username}}</span>
                                <span class="text-muted text-xs block">{{.role}}<b class="caret"></b></span>
                            </a>
                            <ul class="dropdown-menu animated fadeInRight m-t-xs">
                                <li><a class="dropdown-item" href="/layout/account" target="wrapper-content">个人信息</a></li>
                                <li class="dropdown-divider"></li>
                                <li><a class="dropdown-item" href="/logout">退出</a></li>
                            </ul>
                        </div>
                        <div class="logo-element">
                            Dakewe
                        </div>
                    </li>
                    <li class="{{if eq "datainfo" .menu}}active{{end}}">
                        <a href="/layout/datainfo"><i class="fa fa-cubes"></i> <span class="nav-label">编译列表</span></a>
                    </li>
                    <li class="{{if eq "compile" .menu}}active{{end}}">
                        <a href="/layout/compile"><i class="fa fa-align-justify"></i> <span class="nav-label">编译管理</span></a>
                    </li>
                    <li class="{{if eq "knowledge" .menu}}active{{end}}">
                        <a href="/layout/knowledge"><i class="fa fa-book"></i> <span class="nav-label">知识论坛</span></a>
                    </li>
                    <li class="{{if eq "messages" .menu}}active{{end}}">
                        <a href="/layout/messages"><i class="fa fa-weixin"></i> <span class="nav-label">消息管理</span></a>
                    </li>
                    <li class="{{if eq "account" .menu}}active{{end}}">
                        <a href="/layout/account"><i class="fa fa-cogs"></i> <span class="nav-label">账号管理</span></a>
                    </li>
                </ul>

            </div>
        </nav>

        <div id="page-wrapper" class="gray-bg dashbard-1">
            <div class="row border-bottom">
                <nav class="navbar white-bg navbar-static-top" role="navigation" style="margin-bottom: 0">
                    <div class="navbar-header">
                        <a class="navbar-minimalize minimalize-styl-2 btn btn-primary " href="#"><i class="fa fa-bars"></i> </a>
                        <!-- <form role="search" class="navbar-form-custom" action="search_results.html">
                            <div class="form-group">
                                <input type="text" placeholder="请输入需要搜索的内容..." class="form-control" name="top-search" id="top-search">
                            </div>
                        </form> -->
                    </div>
                    <ul class="nav navbar-top-links navbar-right">
                        <li style="padding: 20px">
                            <span class="m-r-sm text-muted welcome-message">欢迎来到{{ .Title }}</span>
                        </li>
                        <li class="dropdown">
                            <a class="dropdown-toggle count-info" data-toggle="dropdown" href="#">
                                <i class="fa fa-bell"></i>  <span class="label label-primary">8</span>
                            </a>
                            <ul class="dropdown-menu dropdown-alerts">
                                <li>
                                    <a href="mailbox.html" class="dropdown-item">
                                        <div>
                                            <i class="fa fa-envelope fa-fw"></i> You have 16 messages
                                            <span class="float-right text-muted small">4 minutes ago</span>
                                        </div>
                                    </a>
                                </li>
                                <li class="dropdown-divider"></li>
                                <li>
                                    <a href="profile.html" class="dropdown-item">
                                        <div>
                                            <i class="fa fa-twitter fa-fw"></i> 3 New Followers
                                            <span class="float-right text-muted small">12 minutes ago</span>
                                        </div>
                                    </a>
                                </li>
                                <li class="dropdown-divider"></li>
                                <li>
                                    <a href="grid_options.html" class="dropdown-item">
                                        <div>
                                            <i class="fa fa-upload fa-fw"></i> Server Rebooted
                                            <span class="float-right text-muted small">4 minutes ago</span>
                                        </div>
                                    </a>
                                </li>
                                <li class="dropdown-divider"></li>
                                <li>
                                    <div class="text-center link-block">
                                        <a href="notifications.html" class="dropdown-item">
                                            <strong>See All Alerts</strong>
                                            <i class="fa fa-angle-right"></i>
                                        </a>
                                    </div>
                                </li>
                            </ul>
                        </li>


                        <li>
                            <a href="/logout">
                                <i class="fa fa-sign-out"></i> 退出
                            </a>
                        </li>
                    </ul>

                </nav>
            </div>
            {{.LayoutContent}}
        </div>
        
        <!-- <div class="row">
            <div class="col-lg-12">
                <div class="wrapper wrapper-content">
                      
                </div>
            </div>
        </div> -->
        <div class="footer">
            <div>
                <strong>Copyright</strong> Dakewe &copy; 2023-2033
            </div>
        </div>
        
    </div>
</body>
</html>

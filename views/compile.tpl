<div class="row">
    <div class="col-lg-12">
        <div class="ibox">
            <div class="ibox-title">
                <h5>
                    编译信息
                </h5>
            </div>
            <div class="ibox-content">
                <form method="post" onsubmit="return false">
                    <div class="form-group row">
                        <label class="col-sm-2 col-form-label">MDK版本</label>
                        <div class="col-sm-10">
                            <select class="form-control m-b" name="mdkVersion">
                                <option value="1">MDK5</option>
                                <option value="2">MDK4</option>
                            </select>
                        </div>
                    </div>
                    <div class="form-group row">
                        <label class="col-sm-2 col-form-label">仓库链接</label>
                        <div class="col-sm-10">
                            <input type="text" class="form-control" name="urlPath">
                        </div>
                    </div>
                    <div class="form-group row">
                        <label class="col-sm-2 col-form-label">Commit</label>
                        <div class="col-sm-10">
                            <input type="text" class="form-control" name="commitValue">
                        </div>
                    </div>
                    <div class="form-group row">
                        <label class="col-sm-2 col-form-label">工程文件名</label>
                        <div class="col-sm-10">
                            <input type="text" class="form-control" name="projectName">
                        </div>
                    </div>
                    <div class="form-group row">
                        <div class="col-sm-4 col-sm-offset-2">
                            <button class="btn btn-white btn-sm" type="reset">取消</button>
                            <button class="btn btn-primary btn-sm" onclick="Compile()">编译</button>
                        </div>
                    </div>
                </form>
            </div>
        </div>
    </div>
    <div class="col-lg-12">
        <div class="ibox">
            <div class="ibox-title">
                <h5>logs</h5>
            </div>
            <div class="ibox-content">
                <div class="md-editor">
                    <textarea name="content" data-provide="markdown" rows="30" class="md-input" style="resize: none;width: 100%;">

                    </textarea>
                </div>
            </div>
        </div>
    </div>
</div>

<script>
    function Compile(){
        $.ajax({
            url: '/lists/commit',
            type: 'POST',
            data: $("form").serializeArray(),
            success: function(res) {
                console.log(res);
            }
        });
        return false;
    }

    $(function() {
        var ws = new WebSocket('ws://'+window.location.host+'/layout/compilews');
        ws.onmessage = function(e) {
            $("textarea").val(e.data);
        }
    })

</script>
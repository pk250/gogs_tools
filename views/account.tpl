<div class="wrapper wrapper-content">
<div class="row">
  <div class="col-lg-6 col-lg-offset-3">
    <div class="ibox">
      <div class="ibox-title"><h5>个人信息</h5></div>
      <div class="ibox-content">
        <form id="account-form">
          <div class="form-group">
            <label>用户名</label>
            <input type="text" class="form-control" value="{{.username}}" disabled>
          </div>
          <div class="form-group">
            <label>邮箱</label>
            <input type="email" class="form-control" id="email" name="email" value="{{.email}}">
          </div>
          <div class="form-group">
            <label>新密码</label>
            <input type="password" class="form-control" id="new_password" name="new_password" placeholder="留空则不修改">
          </div>
          <div class="form-group">
            <label>确认新密码</label>
            <input type="password" class="form-control" id="confirm_password" name="confirm_password" placeholder="留空则不修改">
          </div>
          <button type="submit" class="btn btn-primary">保存</button>
          <span id="save-msg" style="margin-left:12px;display:none;"></span>
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
$('#account-form').submit(function(e) {
    e.preventDefault();
    var newPwd = $('#new_password').val();
    var confirmPwd = $('#confirm_password').val();
    if (newPwd && newPwd !== confirmPwd) {
        toastr.error('两次输入的密码不一致');
        return;
    }
    $.post('/account', $(this).serialize(), function(res) {
        if (res.code === 0) {
            toastr.success('保存成功');
            $('#new_password').val('');
            $('#confirm_password').val('');
        } else {
            toastr.error(res.message || '保存失败');
        }
    });
});
</script>

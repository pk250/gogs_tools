<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>嵌入式CI/CD平台</title>
    <script src="https://cdn.bootcss.com/jquery/3.4.1/jquery.min.js"></script>
    <style>
        * {
            margin: 0;
            padding: 0;
        }
        html {
            height: 100%;
        }
        body {
            height: 100%;
        }
        .container {
            height: 100%;
            background-image: linear-gradient(to right, #fbc2eb, #a6c1ee);
        }
        .login-wrapper {
            background-color: #fff;
            width: 358px;
            height: 588px;
            border-radius: 15px;
            padding: 0 50px;
            position: relative;
            left: 50%;
            top: 50%;
            transform: translate(-50%, -50%);
        }
        .header {
            font-size: 38px;
            font-weight: bold;
            text-align: center;
            line-height: 200px;
        }
        .input-item {
            display: block;
            width: 100%;
            margin-bottom: 20px;
            border: 0;
            padding: 10px;
            border-bottom: 1px solid rgb(128, 125, 125);
            font-size: 15px;
            outline: none;
        }
        .input-item:placeholder {
            text-transform: uppercase;
        }
        .btn {
            text-align: center;
            padding: 10px;
            width: 100%;
            margin-top: 40px;
            background-image: linear-gradient(to right, #a6c1ee, #fbc2eb);
            color: #fff;
        }
        .msg {
            text-align: center;
            line-height: 88px;
        }
        a {
            text-decoration-line: none;
            color: #abc1ee;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="login-wrapper">
            <div class="header">注册</div>
            <div class="form-wrapper">
                <input type="text" name="username" placeholder="username" class="input-item">
                <input type="password" name="password" placeholder="password" class="input-item">
                <input type="text" name="email" placeholder="email" class="input-item">
                <div class="btn">注册</div>
            </div>
            <div class="msg">
                Don't have account?
                <a href="/login">Login</a>
            </div>
        </div>
    </div>
    <script>
        $(function() {
            $('.btn').click(function() {
                var username = $('input[name="username"]').val();
                var password = $('input[name="password"]').val();
                var email = $('input[name="email"]').val();
                $.ajax({
                    url: '/register',
                    type: 'POST',
                    data: {
                        username: username,
                        password: password,
                        email: email
                    },
                    success: function(res) {
                        if (res.status == 1) {
                            window.location.href = res.url;
                        } else {
                            alert(res.msg);
                        }
                    }
                });
            });
        });
        $('body').keydown(function() {
            if (event.keyCode == "13") {
                $('.btn').click();
            }
        });
    </script>
</body>
</html>
## V2GIT

v2git 是一个 HTTP/SSH git服务器。

同时带验证，路由等功能

### 配置

```json
{
  "http_port": 2000,                                // http 服务端口
  "ssh_port": 2222,                                 // ssh 服务端口
  "git_bin_path": "/usr/bin/git",                   // git 路径
  "git_user": "git",                                // 限制ssh的用户名
  "repo_dir": "/Users/repos",                       // 仓库根目录
  "auth_url": "http://127.0.0.1/api/auth",          // 授权URL
  "private_key_path": "/home/git/.ssh/id_rsa"       // 私有key地址
}
```

### 验证

验证时会请求 auth_url。

请求：
```json
{
    "path":"moli/hello.git",

    // SSH
    "fingerprint": "xx:xx:xx:xx……",     // ssh验时的用户公有key指纹

    // HTTP
    "username": "moli",
    "password": "123",
}
```

成功
```json
{
    "result": 0
}
```

失败
```json
{
    "result": 1  //Authentication Failed
}
```
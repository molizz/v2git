## V2GIT

v2git 是 HTTP git服务器。

同时带验证，路由等功能

### 配置

```yml
port: 2000
repobase: /usr/moli/projects         // 仓库root目录
authurl: http://127.0.0.1/api/auth   // 授权时请求的url
```

### 验证

验证时会请求 auth_url。

请求：
```json
{
    "path":"moli/hello.git",
    "username": "moli",
    "password": "123"
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
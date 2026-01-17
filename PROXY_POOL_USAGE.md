# 代理池功能使用说明

## 概述

该功能允许您为不同的OAuth账号配置不同的代理,避免所有账号使用同一个代理导致的风控问题。

## 功能特性

1. **代理池管理** - 在配置文件中定义多个命名代理
2. **账号级代理绑定** - 每个OAuth账号可以绑定特定的代理
3. **Management API** - 通过REST API管理代理池
4. **灵活的代理优先级** - 支持代理池、直接代理URL和全局代理的优先级机制

## 配置方法

### 1. 在config.yaml中定义代理池

```yaml
# 代理池配置
proxy-pool:
  - id: "us-proxy-1"                    # 唯一标识符
    name: "美国代理服务器1"              # 可选的人类可读名称
    proxy-url: "socks5://user:pass@192.168.1.1:1080"
    description: "美国的SOCKS5代理"     # 可选的描述

  - id: "jp-http-proxy"
    name: "日本HTTP代理"
    proxy-url: "http://proxy.jp.example.com:8080"
    description: "日本的HTTP代理"

  - id: "eu-proxy"
    proxy-url: "https://user:pass@eu.proxy.example.com:3128"
```

### 2. OAuth登录时选择代理

目前需要通过直接修改生成的认证文件来绑定代理。未来版本会在登录时提供交互式代理选择。

#### 方法1: 手动编辑认证文件

登录后,在 `~/.cli-proxy-api/` 目录下找到对应的认证文件(如`gemini_xxx.json`, `claude_xxx.json`等),添加`proxy_pool_id`字段:

```json
{
  "id": "xxx",
  "provider": "claude",
  "proxy_pool_id": "us-proxy-1",  // 添加这一行
  "metadata": {
    ...
  }
}
```

#### 方法2: 使用Management API

通过Management API更新账号的代理绑定(需要实现auth文件更新接口)。

### 3. 使用Management API管理代理池

所有Management API都需要认证,在请求头中添加:
```
Authorization: Bearer <your-management-key>
```
或
```
X-Management-Key: <your-management-key>
```

#### 获取所有代理

```bash
GET /v0/management/proxy-pool

curl -H "Authorization: Bearer <key>" http://localhost:8317/v0/management/proxy-pool
```

响应:
```json
{
  "proxy_pool": [
    {
      "id": "us-proxy-1",
      "name": "美国代理服务器1",
      "proxy-url": "socks5://user:pass@192.168.1.1:1080",
      "description": "美国的SOCKS5代理"
    }
  ]
}
```

#### 获取特定代理

```bash
GET /v0/management/proxy-pool/:id

curl -H "Authorization: Bearer <key>" http://localhost:8317/v0/management/proxy-pool/us-proxy-1
```

#### 创建新代理

```bash
POST /v0/management/proxy-pool

curl -X POST -H "Authorization: Bearer <key>" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "new-proxy",
    "name": "新代理",
    "proxy-url": "socks5://127.0.0.1:1080",
    "description": "本地测试代理"
  }' \
  http://localhost:8317/v0/management/proxy-pool
```

#### 更新代理

```bash
PUT /v0/management/proxy-pool/:id

curl -X PUT -H "Authorization: Bearer <key>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "更新后的名称",
    "proxy-url": "socks5://127.0.0.1:1081",
    "description": "更新后的描述"
  }' \
  http://localhost:8317/v0/management/proxy-pool/new-proxy
```

#### 删除代理

```bash
DELETE /v0/management/proxy-pool/:id

curl -X DELETE -H "Authorization: Bearer <key>" \
  http://localhost:8317/v0/management/proxy-pool/new-proxy
```

## 代理优先级

系统按以下优先级选择代理:

1. **代理池** - 如果账号设置了`proxy_pool_id`,使用对应的代理
2. **直接代理URL** - 如果账号设置了`proxy_url`,使用该代理
3. **全局代理** - 使用config.yaml中的`proxy-url`
4. **无代理** - 如果以上都未设置,直连

## 支持的代理协议

- **SOCKS5**: `socks5://[user:pass@]host:port`
- **HTTP**: `http://[user:pass@]host:port`
- **HTTPS**: `https://[user:pass@]host:port`

## 注意事项

1. 代理池ID必须唯一
2. 删除代理池中的代理不会自动更新已绑定该代理的账号
3. 如果账号绑定的代理ID不存在,系统会记录警告并回退到其他代理选项
4. 修改代理池配置后无需重启服务,配置会自动重新加载

## 示例场景

### 场景1: 多地区账号管理

```yaml
proxy-pool:
  - id: "us-west"
    proxy-url: "socks5://us-west.proxy.com:1080"
  - id: "us-east"
    proxy-url: "socks5://us-east.proxy.com:1080"
  - id: "europe"
    proxy-url: "socks5://eu.proxy.com:1080"
  - id: "asia"
    proxy-url: "socks5://asia.proxy.com:1080"
```

然后将不同地区的账号绑定到对应的代理。

### 场景2: 负载均衡

```yaml
proxy-pool:
  - id: "pool-1"
    proxy-url: "socks5://pool1.example.com:1080"
  - id: "pool-2"
    proxy-url: "socks5://pool2.example.com:1080"
  - id: "pool-3"
    proxy-url: "socks5://pool3.example.com:1080"
```

将账号均匀分配到不同的代理池,实现负载均衡。

## 验证代理使用情况

### 方法1: 查看日志 (推荐)

在 `config.yaml` 中启用debug模式：
```yaml
debug: true
```

重启服务后，每次请求都会在日志中显示使用的代理：
```
[DEBUG] Using proxy from pool 'us-proxy-1' for auth abc123
```

### 方法2: 使用Management API查询

#### 查询所有账号的代理信息

```bash
GET /v0/management/auth/proxy-list

curl -H "Authorization: Bearer <key>" \
  http://localhost:8317/v0/management/auth/proxy-list
```

响应示例：
```json
{
  "auths": [
    {
      "auth_id": "abc123",
      "provider": "claude",
      "proxy_pool_id": "us-proxy-1",
      "status": "active",
      "auth_type": "oauth",
      "account": "user@example.com",
      "proxy_info": "via proxy pool: us-proxy-1",
      "resolved_proxy_url": "socks5://user:****@192.168.1.1:1080"
    },
    {
      "auth_id": "def456",
      "provider": "gemini-cli",
      "proxy_pool_id": "",
      "status": "active",
      "auth_type": "oauth",
      "account": "another@example.com (project-123)",
      "proxy_info": "no proxy configured (using global proxy or direct connection)"
    }
  ],
  "total": 2
}
```

#### 查询特定账号的代理信息

```bash
GET /v0/management/auth/:id/proxy

curl -H "Authorization: Bearer <key>" \
  http://localhost:8317/v0/management/auth/abc123/proxy
```

响应示例：
```json
{
  "auth_id": "abc123",
  "provider": "claude",
  "proxy_pool_id": "us-proxy-1",
  "proxy_url": "",
  "resolved_proxy_url": "socks5://user:pass@192.168.1.1:1080",
  "proxy_info": "via proxy pool: us-proxy-1",
  "auth_type": "oauth",
  "account": "user@example.com"
}
```

如果代理池ID不存在，会有警告：
```json
{
  "auth_id": "abc123",
  "provider": "claude",
  "proxy_pool_id": "non-existent",
  "resolved_proxy_url": "",
  "warning": "proxy pool ID not found in configuration",
  "proxy_info": "via proxy pool: non-existent"
}
```

### 方法3: 查看认证文件

```bash
cat ~/.cli-proxy-api/claude_abc123.json
```

查看 `proxy_pool_id` 字段：
```json
{
  "id": "abc123",
  "provider": "claude",
  "proxy_pool_id": "us-proxy-1",
  "metadata": {
    "email": "user@example.com"
  }
}
```

### 方法4: 测试真实请求

发送一个测试请求，通过代理服务器的访问日志验证：

```bash
curl -X POST http://localhost:8317/v1/chat/completions \
  -H "Authorization: Bearer <your-api-key>" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-sonnet-4",
    "messages": [{"role": "user", "content": "Hello"}]
  }'
```

然后查看代理服务器的日志，确认请求来源。

## 故障排除

### 代理连接失败

检查日志中的错误信息:
```
Using proxy from pool 'xxx' for auth yyy
```

如果看到警告:
```
Proxy pool ID 'xxx' not found in configuration for auth yyy
```

说明代理ID不存在,请检查:
1. config.yaml中是否定义了该代理
2. 代理ID拼写是否正确

### 代理不生效

确认代理优先级,查看日志中使用的代理信息。

## 未来改进

- [ ] 在OAuth登录时提供交互式代理选择
- [ ] 通过CLI命令管理代理池
- [ ] 通过Management API更新账号的代理绑定
- [ ] 代理健康检查和自动故障转移
- [ ] 代理使用统计和监控

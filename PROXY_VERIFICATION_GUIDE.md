# 代理池功能快速验证指南

## 快速验证步骤

### 1. 启用Debug日志

编辑 `config.yaml`：
```yaml
debug: true
```

### 2. 配置代理池

在 `config.yaml` 中添加：
```yaml
proxy-pool:
  - id: "test-proxy"
    name: "测试代理"
    proxy-url: "socks5://127.0.0.1:1080"  # 替换为你的代理地址
    description: "用于测试的代理"
```

### 3. 为账号绑定代理

编辑你的认证文件 `~/.cli-proxy-api/<provider>_<id>.json`，添加：
```json
{
  "id": "...",
  "provider": "...",
  "proxy_pool_id": "test-proxy",  // 添加这一行
  ...
}
```

### 4. 重启服务

```bash
./cli-proxy-api run
```

### 5. 验证代理配置

#### 方法A: 查看所有账号的代理 (推荐)

```bash
curl -H "Authorization: Bearer <your-management-key>" \
  http://localhost:8317/v0/management/auth/proxy-list
```

你会看到类似输出：
```json
{
  "auths": [
    {
      "auth_id": "xxx",
      "provider": "claude",
      "proxy_pool_id": "test-proxy",
      "proxy_info": "via proxy pool: test-proxy",
      "resolved_proxy_url": "socks5://****@127.0.0.1:1080",
      "account": "your@email.com"
    }
  ]
}
```

✅ 如果看到 `"proxy_pool_id": "test-proxy"` 和 `"resolved_proxy_url"` 有值，说明配置成功！

#### 方法B: 查看特定账号

```bash
# 先获取账号ID
AUTH_ID=$(cat ~/.cli-proxy-api/*.json | grep -m 1 '"id"' | cut -d'"' -f4)

# 查询该账号的代理信息
curl -H "Authorization: Bearer <your-management-key>" \
  "http://localhost:8317/v0/management/auth/${AUTH_ID}/proxy"
```

### 6. 测试请求并查看日志

发送一个测试请求：
```bash
curl -X POST http://localhost:8317/v1/chat/completions \
  -H "Authorization: Bearer <your-api-key>" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-sonnet-4",
    "messages": [{"role": "user", "content": "test"}],
    "max_tokens": 10
  }'
```

查看服务日志，应该能看到：
```
[DEBUG] Using proxy from pool 'test-proxy' for auth xxx
```

## 常见问题排查

### ❌ 问题1: API返回 "proxy pool ID not found"

**原因**: 代理池ID拼写错误或未在config.yaml中定义

**解决**:
1. 检查 `config.yaml` 中的 `proxy-pool` 配置
2. 确认ID完全匹配（区分大小写）
3. 重启服务以加载新配置

### ❌ 问题2: 日志中没有代理信息

**原因**: debug模式未启用

**解决**:
```yaml
debug: true
```

### ❌ 问题3: 代理不生效

**检查步骤**:
1. 确认 `proxy_pool_id` 已保存到认证文件
2. 确认代理服务器正在运行
3. 使用Management API验证配置：
   ```bash
   curl -H "Authorization: Bearer <key>" \
     http://localhost:8317/v0/management/auth/proxy-list
   ```
4. 查看日志确认是否使用了代理

### ❌ 问题4: 找不到auth ID

**获取所有auth ID的方法**:
```bash
# Linux/Mac
ls ~/.cli-proxy-api/*.json | xargs -I {} basename {} .json

# 或通过API
curl -H "Authorization: Bearer <key>" \
  http://localhost:8317/v0/management/auth/proxy-list | jq '.auths[].auth_id'
```

## 验证成功的标志

✅ Management API显示正确的 `proxy_pool_id` 和 `resolved_proxy_url`
✅ 日志中显示 `Using proxy from pool 'xxx' for auth yyy`
✅ 认证文件中有 `"proxy_pool_id": "xxx"` 字段
✅ 代理服务器日志显示来自CLIProxyAPI的连接

## 下一步

配置成功后，你可以：
- 为不同账号配置不同代理
- 使用Management API动态管理代理池
- 监控各个代理的使用情况
- 实现基于地区的代理分配策略

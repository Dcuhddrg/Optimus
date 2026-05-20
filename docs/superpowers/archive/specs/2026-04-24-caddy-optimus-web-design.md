# Caddy HTTPS 代理 Optimus-web 设计

## 目标

通过宿主机 Caddy 为 Optimus-web 提供 HTTPS 域名访问 `https://optimus.intellivis.cn`。

## 架构

```
Client (HTTPS) → optimus.intellivis.cn (DNS → 服务器 IP)
  → Caddy (宿主机 :443) → localhost:18080 (Optimus-web Nginx)
```

## 配置

### Caddyfile

```
optimus.intellivis.cn {
    reverse_proxy localhost:18080
}
```

Caddy 自动处理：
- TLS 证书（Let's Encrypt 自动申请与续期）
- HTTP → HTTPS 重定向
- 反向代理到 `localhost:18080`

### 前提条件

1. `optimus.intellivis.cn` 公网 DNS A 记录指向服务器 IP
2. 宿主机防火墙放行 80 和 443 端口
3. optimus-web 容器运行且 18080 端口可访问

## 不需要的

- 修改 `docker-compose.yml` 或网络
- 代理后端 API（Nginx 已在 `/api/` 路径处理）
- Caddy 容器化

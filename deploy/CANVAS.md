# canvas.xlabapi.com 部署说明

## 概述

Infinite Canvas 作为独立容器运行，通过 Caddy 反代至 `canvas.xlabapi.com`。

- 容器镜像：`ghcr.io/basketikun/infinite-canvas:latest`
- 容器内端口：3000，宿主机映射：`127.0.0.1:3001`
- 用户配置存浏览器本地存储，无需任何服务端环境变量

## 部署步骤

### 1. DNS

在域名服务商处为 `canvas.xlabapi.com` 添加 A 记录，指向服务器 IP：

```
canvas.xlabapi.com.  A  152.53.39.161
```

### 2. 启动容器

```bash
cd /path/to/deploy
docker compose up -d infinite-canvas
```

验证容器正常运行：

```bash
docker compose ps infinite-canvas
curl -s -o /dev/null -w "%{http_code}" http://127.0.0.1:3001/
```

### 3. 重载 Caddy

```bash
docker exec caddy caddy reload --config /etc/caddy/Caddyfile
# 或宿主机直接执行（视部署方式而定）
caddy reload --config /etc/caddy/Caddyfile
```

Caddy 会自动申请并续签 TLS 证书（Let's Encrypt）。

## 注意事项

- 端口 3001 仅绑定 `127.0.0.1`，不对外暴露，所有流量经由 Caddy 转发
- 镜像来源：`ghcr.io/basketikun/infinite-canvas`（GitHub Container Registry）
- 若 ghcr.io 拉取失败，可改为本地 build 方式：将 `image:` 替换为 `build: { context: ./infinite-canvas, dockerfile: Dockerfile }`

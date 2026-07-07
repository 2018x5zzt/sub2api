# canvas.xlabapi.com 部署说明

## 概述

`ImageStudioView.vue` 只是一个 iframe 壳，真实内容来自 `https://canvas.xlabapi.com`。线上部署有两个前提必须同时成立：

- `canvas.xlabapi.com` 需要由反向代理转发到本机 `127.0.0.1:13000`
- 主站 CSP 的 `frame-src` 需要允许 `https://canvas.xlabapi.com`

当前宝塔线上形态以 `nginx + docker` 为准，不是 Caddy。

- 容器镜像：`ghcr.io/basketikun/infinite-canvas:latest`
- 容器内端口：`3000`
- 宿主机映射：`0.0.0.0:13000 -> 3000`
- TLS 证书：直接复用 `xlabapi.com` 的通配证书（含 `*.xlabapi.com`）

## 部署步骤

### 1. DNS

确认 `canvas.xlabapi.com` 已解析到线上服务器：

```text
canvas.xlabapi.com.  A  152.53.39.161
```

### 2. 启动容器

```bash
docker ps --format 'table {{.Names}}\t{{.Ports}}' | grep infinite-canvas
docker start infinite-canvas
curl -sI http://127.0.0.1:13000
```

预期 `curl` 返回 `200 OK`，并且响应头带有 Next.js 相关字段，例如 `x-nextjs-cache`。

### 3. 配置宝塔 Nginx

`/www/server/panel/vhost/nginx/canvas.xlabapi.com.conf` 需要同时覆盖 `80` 和 `443`：

```nginx
server {
    listen 80;
    listen [::]:80;
    server_name canvas.xlabapi.com;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name canvas.xlabapi.com;

    ssl_certificate     /www/server/panel/vhost/cert/xlabapi.com/fullchain.pem;
    ssl_certificate_key /www/server/panel/vhost/cert/xlabapi.com/privkey.pem;

    location ^~ / {
        proxy_pass http://127.0.0.1:13000;
        proxy_set_header Host $http_host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection $connection_upgrade;
        proxy_buffering off;
    }
}
```

重载：

```bash
nginx -t
nginx -s reload
```

### 4. 主站校验

修完反代后，再确认主站允许嵌入：

```bash
curl -sI https://xlabapi.com/image-studio | grep -i content-security-policy
```

响应头里的 `frame-src` 必须包含 `https://canvas.xlabapi.com`。

## 常见故障

- `canvas.xlabapi.com` 返回 `404 Not Found` 且 `server: nginx`
  说明 HTTPS 请求落到了宝塔默认站点，通常是没配 `443 server_name` 或没挂正确证书。
- `curl https://canvas.xlabapi.com` 报证书错误
  说明命中了默认自签证书，还是 HTTPS vhost 没接住。
- `/image-studio` 页面空白，但新窗口能打开 `canvas`
  说明主站 CSP 没放行 `https://canvas.xlabapi.com`。

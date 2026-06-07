# Xlab Backend Docker Phase 2 设计

## 背景

Phase 1 已经建立了 xlab shell 的最小代码边界：`xlab-backend` 可以通过 core JWT 验证用户，并代理产品订阅只读接口；`frontend-v2` 已经具备 `/xapi/v1` adapter。当前生产部署仍可使用兼容模式 `VITE_XLAB_API_BASE_URL=/api/v1`，让产品订阅页继续调用 core 的 `/api/v1/subscription-products/*`。

Phase 2 的目标是让生产真正运行独立 `xlab-backend` 容器，并通过同源 `/xapi/v1` 反向代理访问它。此阶段仍然不迁移数据、不迁移支付履约、不删除 core 产品订阅逻辑。

## 目标

1. 为 `xlab-backend` 增加 Docker 构建能力。
2. 增加部署脚本，能在生产服务器上创建或更新 `xlab-backend` 容器。
3. 明确 `/xapi/v1 -> xlab-backend:8090` 的反向代理要求和探测步骤。
4. 将 frontend-v2 从兼容模式切换到 `/xapi/v1` 前提供安全验证门禁。
5. 保留快速回滚到 `/api/v1` 的能力。

## 非目标

- 不迁移产品订阅数据。
- 不迁移支付订单、支付回调或 payment fulfillment。
- 不迁移联盟、兑换码商品化或权益同步逻辑。
- 不删除 core 中的 `subscription-products` API。
- 不要求本阶段引入完整 docker-compose 编排；先使用部署脚本管理单独容器。

## 目标拓扑

```text
Browser
  |
  | same origin
  v
Reverse proxy / ingress
  |-- /api/v1   -> sub2api:8080
  |-- /v1       -> sub2api:8080
  |-- /xapi/v1  -> xlab-backend:8090
  `-- /*        -> sub2api:8080 embedded frontend-v2
```

Docker 侧：

```text
sub2api container
xlab-backend container
same docker network
```

`xlab-backend` 环境变量：

```text
XLAB_SERVER_ADDR=:8090
CORE_API_BASE_URL=http://sub2api:8080/api/v1
XLAB_CORE_TIMEOUT_SECONDS=10
```

## 部署流程

### 1. 部署前探测

先确认服务器当前网络和容器拓扑：

```bash
ssh root@152.53.39.161 "docker ps --format '{{.Names}}\t{{.Status}}'"
ssh root@152.53.39.161 "docker inspect sub2api --format '{{json .NetworkSettings.Networks}}'"
ssh root@152.53.39.161 "ss -lntp | grep -E ':80|:443|:8080|:8090' || true"
```

这一步用于判断 `/xapi/v1` 应配置在外部 Nginx/Caddy，还是需要新增前置代理容器。

### 2. 构建并部署 xlab-backend 容器

部署脚本应：

1. 本地构建 `xlab-backend` Linux binary。
2. 本地构建轻量 Docker image，并保存成 tar。
3. 上传 image tar 到服务器。
4. 识别 `sub2api` 所在 Docker network。
5. 停止并删除旧 `xlab-backend` 容器。
6. 以同一 network 启动新容器。
7. 设置 `CORE_API_BASE_URL=http://sub2api:8080/api/v1`。
8. 映射 `127.0.0.1:8090:8090`，方便宿主机反代到 xlab-backend。
9. 验证宿主机 `/health`。

### 3. 配置 `/xapi/v1` 反向代理

如果线上使用 Nginx，可添加：

```nginx
location /xapi/v1/ {
    proxy_pass http://127.0.0.1:8090/xapi/v1/;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

如果没有外部反代，Phase 2 先只部署 xlab-backend 容器并保持 frontend-v2 兼容模式；不切 `/xapi/v1`。

### 4. 切换 frontend-v2 到 `/xapi/v1`

只有当下面检查通过后才切换：

```bash
curl -i https://<domain>/xapi/v1/subscription-products/active
```

预期未登录返回 JSON 401，而不是 404 或 HTML。

然后重新部署 frontend-v2/core：

```bash
VITE_XLAB_API_BASE_URL=/xapi/v1 bash ./deploy.sh
```

## 回滚方案

如果 `/xapi/v1` 出现问题，不需要改数据库：

1. 重新用兼容模式部署 frontend-v2：

   ```bash
   VITE_XLAB_API_BASE_URL=/api/v1 bash ./deploy.sh
   ```

2. 停止 xlab-backend：

   ```bash
   docker stop xlab-backend
   ```

3. core 的 `/api/v1/subscription-products/*` 保持可用，订阅页回到当前行为。

## 验证标准

- `xlab-backend` 容器运行中。
- 服务器宿主机 `http://127.0.0.1:8090/health` 返回 200。
- `/xapi/v1/subscription-products/active` 未登录返回 JSON 401。
- 登录后 frontend-v2 订阅页正常展示产品订阅。
- `/api/v1/subscription-products/active` 仍可用。
- `frontend-v2` build 时 `VITE_XLAB_API_BASE_URL=/xapi/v1` 可生效。

## 风险

| 风险 | 缓解 |
|---|---|
| 反代未配置导致订阅页 404 | 先验证 `/xapi/v1` 未登录返回 JSON 401，再切 frontend。 |
| xlab-backend 无法访问 core | 将两个容器放到同一 Docker network，使用 `http://sub2api:8080/api/v1`。 |
| 上线后订阅页失败 | 立即以 `VITE_XLAB_API_BASE_URL=/api/v1` 回滚 frontend。 |
| 部署脚本影响现有 core | xlab-backend 独立部署，core deploy 脚本仅在切换 frontend 时执行。 |

## 后续

Phase 2 完成后，才能进入 Phase 3：把产品订阅只读数据源从 core 迁到 xlab DB，并保留 core API 作为回滚备用。

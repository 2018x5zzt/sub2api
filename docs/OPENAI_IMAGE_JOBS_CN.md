# OpenAI 异步生图接口（Image Jobs）

本文档描述 xlabapi 在 OpenAI 兼容层提供的异步生图接口，用于避免长耗时同步请求导致的客户端超时问题。

## 1. 适用场景

- 生图任务耗时较长（几十秒到数分钟）
- 客户端不希望长时间阻塞 HTTP 连接
- 需要“提交任务 + 轮询结果”的稳定调用方式

## 2. 鉴权方式

与现有网关一致，支持以下任一方式：

- `Authorization: Bearer <API_KEY>`
- `x-api-key: <API_KEY>`

## 3. 接口总览

### 3.1 提交生成任务

- `POST /v1/images/jobs/generations`
- 别名：`POST /images/jobs/generations`

请求体与同步接口 `POST /v1/images/generations` 保持一致。

### 3.2 提交编辑任务

- `POST /v1/images/jobs/edits`
- 别名：`POST /images/jobs/edits`

请求体与同步接口 `POST /v1/images/edits` 保持一致。

### 3.3 查询任务状态

- `GET /v1/images/jobs/{job_id}`
- 别名：`GET /images/jobs/{job_id}`

> 轮询接口为查询型请求，会走“鉴权但跳过计费闸门”路径，避免成功扣费后因额度状态变化导致无法取回结果。

## 4. 请求与响应示例

## 4.1 提交生成任务

```bash
curl -sS 'https://YOUR_DOMAIN/v1/images/jobs/generations' \
  -H 'Authorization: Bearer sk-xxx' \
  -H 'Content-Type: application/json' \
  -d '{
    "model": "gpt-image-2",
    "prompt": "A glass cat reading a book",
    "size": "1024x1024"
  }'
```

成功响应（`202 Accepted`）：

```json
{
  "id": "imgjob_5c1f5d7e-0f95-4f7a-b8b7-9d9f8f767b2c",
  "job_id": "imgjob_5c1f5d7e-0f95-4f7a-b8b7-9d9f8f767b2c",
  "status": "queued",
  "created_at": 1747290000,
  "updated_at": 1747290000
}
```

## 4.2 轮询任务状态

```bash
curl -sS 'https://YOUR_DOMAIN/v1/images/jobs/imgjob_5c1f5d7e-0f95-4f7a-b8b7-9d9f8f767b2c' \
  -H 'Authorization: Bearer sk-xxx'
```

运行中（`200 OK`）：

```json
{
  "id": "imgjob_5c1f5d7e-0f95-4f7a-b8b7-9d9f8f767b2c",
  "job_id": "imgjob_5c1f5d7e-0f95-4f7a-b8b7-9d9f8f767b2c",
  "status": "running",
  "created_at": 1747290000,
  "updated_at": 1747290002
}
```

成功完成（`200 OK`）：

```json
{
  "id": "imgjob_5c1f5d7e-0f95-4f7a-b8b7-9d9f8f767b2c",
  "job_id": "imgjob_5c1f5d7e-0f95-4f7a-b8b7-9d9f8f767b2c",
  "status": "succeeded",
  "created_at": 1747290000,
  "updated_at": 1747290018,
  "completed_at": 1747290018,
  "http_status": 200,
  "response": {
    "created": 1747290018,
    "data": [
      { "url": "https://..." }
    ]
  }
}
```

失败完成（`200 OK`，任务状态失败）：

```json
{
  "id": "imgjob_5c1f5d7e-0f95-4f7a-b8b7-9d9f8f767b2c",
  "job_id": "imgjob_5c1f5d7e-0f95-4f7a-b8b7-9d9f8f767b2c",
  "status": "failed",
  "created_at": 1747290000,
  "updated_at": 1747290015,
  "completed_at": 1747290015,
  "http_status": 502,
  "error": {
    "type": "api_error",
    "message": "Image job failed"
  }
}
```

## 5. 状态机

- `queued`：任务已入队，等待执行
- `running`：任务执行中
- `succeeded`：执行成功，可从 `response` 读取上游结果
- `failed`：执行失败，可从 `error` 和 `http_status` 查看失败信息

## 6. 错误码与常见问题

- `404 not_found_error`
  - 平台不是 OpenAI 组
  - `job_id` 不存在
  - `job_id` 不属于当前 API Key 用户
- `400 invalid_request_error`
  - 请求体为空或 JSON 非法
- `401 authentication_error`
  - API Key 无效或未提供
- `413 invalid_request_error`
  - 请求体超过限制

## 7. 轮询建议

- 建议轮询间隔：`2~3 秒`
- 建议最大轮询时长：`20 分钟`（与服务端 job 执行超时默认值一致）
- 客户端在 `status in [succeeded, failed, timeout]` 后停止轮询

## 8. 实现说明（当前版本）

- Job 状态与结果存储在 Redis：
  - `POST` 提交任务的实例仍负责本地执行该任务
  - `GET` 轮询可命中任意 API Server 实例
  - 服务重启后，Redis TTL 未过期的 `job_id` 仍可查询
- Watchdog 会定期扫描 Redis 中 stale 的 `queued/running` 任务，并将其标记为 `timeout`（`http_status: 504`）
- 当前版本不是 Redis 分布式执行队列：
  - 如果执行实例进程崩溃，任务不会被其他实例接手重跑
  - 崩溃遗留的 `queued/running` 任务会由 watchdog 终态化为 `timeout`
- 默认参数：
  - 并发执行槽：`2`
  - 单任务超时：`20m`
  - Job 保留 TTL：`24h`
  - Watchdog 扫描周期：`60s`

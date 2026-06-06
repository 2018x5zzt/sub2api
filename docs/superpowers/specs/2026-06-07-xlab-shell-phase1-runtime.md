# Xlab Shell Phase 1 Runtime Wiring

Phase 1 introduces a standalone xlab backend with read-only product subscription endpoints.

## Services

- sub2api core: existing upstream-compatible core service.
- xlab backend: new service listening on `XLAB_SERVER_ADDR`, default `:8090`.
- frontend-v2: calls `/api/v1` for core and `/xapi/v1` for xlab.

## Required environment

For xlab backend:

```text
XLAB_SERVER_ADDR=:8090
CORE_API_BASE_URL=http://sub2api:8080/api/v1
XLAB_CORE_TIMEOUT_SECONDS=10
```

For frontend-v2 build:

```text
VITE_XLAB_API_BASE_URL=/xapi/v1
```

## Reverse proxy rules

```text
/api/v1  -> sub2api core
/v1      -> sub2api core gateway
/xapi/v1 -> xlab backend
/*       -> frontend-v2 shell
```

## Rollout note

The xlab backend initially proxies current core product subscription reads. It does not migrate product subscription data and does not change payment fulfillment.

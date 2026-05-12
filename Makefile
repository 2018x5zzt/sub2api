.PHONY: build build-backend build-frontend build-datamanagementd test test-backend test-frontend test-datamanagementd secret-scan

# 一键编译前后端
build: build-backend build-frontend

# 编译后端（复用 backend/Makefile）
build-backend:
	@$(MAKE) -C backend build

# 编译前端（当前默认使用 React frontend-v2，旧 Vue frontend 已从默认构建路径下线）
build-frontend:
	@npm --prefix frontend-v2 run build

# 编译 datamanagementd（宿主机数据管理进程）
build-datamanagementd:
	@cd datamanagement && go build -o datamanagementd ./cmd/datamanagementd

# 运行测试（后端 + 前端）
test: test-backend test-frontend

test-backend:
	@$(MAKE) -C backend test

test-frontend:
	@npm --prefix frontend-v2 run typecheck

test-datamanagementd:
	@cd datamanagement && go test ./...

secret-scan:
	@python3 tools/secret_scan.py

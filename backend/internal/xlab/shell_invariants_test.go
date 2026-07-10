//go:build unit

package xlab_test

import (
	"os"
	"strings"
	"testing"
)

// 薄壳硬不变量测试 (H1–H6)
// 每条测试守住一个"曾经因 roll-up 被静默覆盖导致生产宕机"的接缝。
// 失败时说明如何恢复，以便快速定位。
//
// 运行方式：go test -tags=unit ./backend/internal/xlab/...

const (
	repoRoot = "../../.." // 相对于本文件的仓库根目录（backend/internal/xlab → root）
)

func mustReadFile(t *testing.T, relPath string) string {
	t.Helper()
	data, err := os.ReadFile(repoRoot + "/" + relPath)
	if err != nil {
		t.Fatalf("无法读取文件 %s: %v\n恢复方式：确认该文件在 xlabapi 分支上存在，若丢失请从 xlabapi-snapshot-0.1.137 中恢复", relPath, err)
	}
	return string(data)
}

// H1: Miku OAuth provider 及其 Redis code store 存在且包含核心常量。
// 丢失后果：Miku SSO 登录全部返回 404。已丢失过 2 次（0.1.123、0.1.139 重滚）。
// 恢复方式：从 xlabapi-snapshot-0.1.137 中取回 backend/internal/handler/auth_xlab_oauth_provider.go。
func TestH1_XlabOAuthProviderExists(t *testing.T) {
	checks := []struct {
		path   string
		marker string
	}{
		{"backend/internal/handler/auth_xlab_oauth_provider.go", "xlabOAuthProviderClientMiku"},
		{"backend/internal/handler/auth_handler.go", "xlabOAuthCodeStore"},
		{"backend/internal/repository/xlab_oauth_code_store.go", "xlabOAuthProviderCodeKeyPrefix"},
		{"backend/internal/repository/wire.go", "NewXlabOAuthCodeStore"},
		{"backend/cmd/server/wire_gen.go", "xlabOAuthCodeStore := repository.NewXlabOAuthCodeStore"},
	}
	for _, check := range checks {
		content := mustReadFile(t, check.path)
		if !strings.Contains(content, check.marker) {
			t.Errorf("H1 FAIL: %s 缺少标识符 %q\n恢复方式：从 xlabapi-snapshot-0.1.137 恢复对应 xlab OAuth 实现", check.path, check.marker)
		}
	}
}

// H2: embed_on.go 对 /oauth/token 和 /oauth/userinfo 的 bypass 存在。
// 丢失后果：xlab OAuth 颁发 token 的端点被前端静态文件拦截，返回 HTML 而非 JSON。
// 恢复方式：确认 backend/internal/web/embed_on.go 中 shouldBypassEmbed 函数包含这两条路径。
func TestH2_EmbedBypassExists(t *testing.T) {
	content := mustReadFile(t, "backend/internal/web/embed_on.go")
	required := []string{
		`"/oauth/token"`,
		`"/oauth/userinfo"`,
	}
	for _, r := range required {
		if !strings.Contains(content, r) {
			t.Errorf("H2 FAIL: embed_on.go 缺少 bypass 路径 %s\n恢复方式：在 shouldBypassEmbed 函数中加回该路径判断", r)
		}
	}
}

// H3: migrations_runner.go 包含全部4条 xlabapi 专有 checksum 兼容规则。
// 丢失后果：生产启动时迁移 checksum 校验失败，服务无法启动（已发生过2次）。
// 恢复方式：在 migrationChecksumCompatibilityRules map 中补回对应条目。
func TestH3_MigrationCompatRulesExist(t *testing.T) {
	content := mustReadFile(t, "backend/internal/repository/migrations_runner.go")
	required := []string{
		"047_add_sora_pricing_and_media_type.sql",
		"063_add_sora_client_tables.sql",
		"107_add_account_cost_to_dashboard_tables.sql",
		"140_restore_shared_subscription_products.sql",
	}
	for _, r := range required {
		if !strings.Contains(content, r) {
			t.Errorf("H3 FAIL: migrations_runner.go 缺少 checksum 兼容规则 %q\n恢复方式：在 migrationChecksumCompatibilityRules 中补回该条目（参考 xlabapi-snapshot-0.1.137）", r)
		}
	}
}

// H4: gateway_handler.go 中 productSettlement 被传入结算链路。
// 丢失后果：product-subscription 用户按量计费失效，扣费不走订阅通道。
// 恢复方式：确认 gateway_handler.go 中 productSettlement 变量被声明并传入请求上下文。
func TestH4_ProductSettlementInGateway(t *testing.T) {
	content := mustReadFile(t, "backend/internal/handler/gateway_handler.go")
	if !strings.Contains(content, "productSettlement") {
		t.Errorf("H4 FAIL: gateway_handler.go 中找不到 productSettlement\n恢复方式：确认 product-subscription 结算上下文注入代码未被上游覆盖")
	}
}

// H5: 根 Dockerfile 和 deploy/Dockerfile 均包含 COPY docs/legal/。
// 丢失后果：前端 Docker 构建失败（AdminComplianceDialog.vue import *.md?raw 找不到文件）。
// 恢复方式：在两个 Dockerfile 的前端构建阶段加回 COPY docs/legal/ /app/docs/legal/。
func TestH5_DockerfileCopyDocsLegal(t *testing.T) {
	for _, path := range []string{"Dockerfile", "deploy/Dockerfile"} {
		content := mustReadFile(t, path)
		if !strings.Contains(content, "COPY docs/legal/") {
			t.Errorf("H5 FAIL: %s 缺少 COPY docs/legal/\n恢复方式：在前端构建阶段加回该指令", path)
		}
	}
}

// H6: 根 Dockerfile 和 deploy/Dockerfile 均锁定 pnpm@9。
// 丢失后果：构建时拉取 pnpm@latest（pnpm 11），触发 onlyBuiltDependencies 问题导致前端构建失败。
// 恢复方式：将两个 Dockerfile 中的 corepack prepare pnpm@latest 改回 pnpm@9。
func TestH6_DockerfilePnpm9(t *testing.T) {
	for _, path := range []string{"Dockerfile", "deploy/Dockerfile"} {
		content := mustReadFile(t, path)
		if !strings.Contains(content, "pnpm@9") {
			t.Errorf("H6 FAIL: %s 未锁定 pnpm@9\n恢复方式：将 corepack prepare 版本改为 pnpm@9 --activate", path)
		}
	}
}

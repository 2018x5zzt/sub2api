//go:build unit

package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestShouldFailoverOn400_OfficialCapacityLimitPlainText(t *testing.T) {
	svc := &GatewayService{}

	require.True(t, svc.shouldFailoverOn400([]byte("官方算力限制，请等待一段时间后再进行使用，如有问题可联系管理员")))
}

func TestShouldFailoverOn400_InvalidRequestDoesNotFailover(t *testing.T) {
	svc := &GatewayService{}

	require.False(t, svc.shouldFailoverOn400([]byte(`{"error":{"message":"messages.0.content.0.text: field required"}}`)))
}

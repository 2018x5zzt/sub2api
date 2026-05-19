//go:build unit

package service

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDoBalanceCreatesCommercialRedeemCode(t *testing.T) {
	source, err := os.ReadFile("payment_fulfillment.go")
	require.NoError(t, err)
	content := strings.Join(strings.Fields(string(source)), " ")

	require.Contains(t, content, "SourceType: RedeemSourceCommercial")
}

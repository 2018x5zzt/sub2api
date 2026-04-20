//go:build unit

package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseEnterpriseVisibleGroupIDsByEnterprise_NormalizesEnterpriseNames(t *testing.T) {
	got, err := ParseEnterpriseVisibleGroupIDsByEnterprise(`{" Acme ":[9,2,9],"acme":[11],"Bustest":[5]}`)

	require.NoError(t, err)
	require.Equal(t, map[string][]int64{
		"acme":    {2, 9, 11},
		"bustest": {5},
	}, got)
}

func TestBuildEnterpriseVisibleGroupRules_SortsAndDeduplicates(t *testing.T) {
	got := BuildEnterpriseVisibleGroupRules(map[string][]int64{
		"bustest": {18, 16, 16},
		"acme":    {9, 2},
	})

	require.Equal(t, []EnterpriseVisibleGroupSetting{
		{
			EnterpriseName:  "acme",
			VisibleGroupIDs: []int64{2, 9},
		},
		{
			EnterpriseName:  "bustest",
			VisibleGroupIDs: []int64{16, 18},
		},
	}, got)
}

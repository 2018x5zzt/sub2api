package enterprisebff

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseEnterpriseVisibleGroupIDsByEnterprise_ParsesJSONMapping(t *testing.T) {
	got, err := parseEnterpriseVisibleGroupIDsByEnterprise(`{"bustest":[2,9,11],"Acme":[5]}`)

	require.NoError(t, err)
	require.Equal(t, map[string][]int64{
		"bustest": {2, 9, 11},
		"acme":    {5},
	}, got)
}

func TestParseEnterpriseVisibleGroupIDsByEnterprise_RejectsInvalidJSON(t *testing.T) {
	_, err := parseEnterpriseVisibleGroupIDsByEnterprise(`{"bustest":"not-an-array"}`)

	require.Error(t, err)
}

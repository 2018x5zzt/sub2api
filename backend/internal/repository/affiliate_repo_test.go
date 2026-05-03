package repository

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestCountEffectiveInviteesUsesPaidPaymentOrders(t *testing.T) {
	db, mock := newSQLMock(t)

	mock.ExpectQuery("payment_orders po").
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(52)))

	count, err := countEffectiveInvitees(context.Background(), db, 7)

	require.NoError(t, err)
	require.Equal(t, int64(52), count)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCountEffectiveInviteesIncludesCommercialAndSubscriptionRedeems(t *testing.T) {
	db, mock := newSQLMock(t)

	mock.ExpectQuery("redeem_codes rc").
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(64)))

	count, err := countEffectiveInvitees(context.Background(), db, 7)

	require.NoError(t, err)
	require.Equal(t, int64(64), count)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAffiliateUserOverviewSQLIncludesMaturedFrozenQuota(t *testing.T) {
	query := strings.Join(strings.Fields(affiliateUserOverviewSQL), " ")

	require.Contains(t, query, "ua.aff_quota + COALESCE(matured.matured_frozen_quota, 0)")
	require.Contains(t, query, "frozen_until <= NOW()")
}

func TestAffiliateRecordQueriesUseLedgerAuditFields(t *testing.T) {
	source, err := os.ReadFile("affiliate_repo.go")
	require.NoError(t, err)
	content := string(source)

	require.Contains(t, content, "JOIN payment_orders po ON po.id = ual.source_order_id")
	require.Contains(t, content, "ual.amount::double precision")
	require.Contains(t, content, "ual.balance_after::double precision")
	require.NotContains(t, content, "parseAffiliateRebateAmount")
	require.NotContains(t, content, `"current_balance": "u.balance"`)
}

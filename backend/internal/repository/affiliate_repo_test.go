package repository

import (
	"context"
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

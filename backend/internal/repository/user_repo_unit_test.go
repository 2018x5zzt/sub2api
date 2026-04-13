//go:build unit

package repository

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestNormalizedInviteCodePreservesCase(t *testing.T) {
	require.Equal(t, "AbCdEfGh", normalizedInviteCode("  AbCdEfGh  "))
}

func TestFindInviteAliasUserIDReturnsMappedUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"user_id"}).AddRow(int64(42))
	mock.ExpectQuery("SELECT user_id FROM invite_code_aliases").
		WithArgs("U1").
		WillReturnRows(rows)

	repo := newUserRepositoryWithSQL(nil, db)
	userID, err := repo.findInviteAliasUserID(context.Background(), "U1")
	require.NoError(t, err)
	require.EqualValues(t, 42, userID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestInviteAliasExistsReturnsTrue(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"exists"}).AddRow(true)
	mock.ExpectQuery("SELECT EXISTS").
		WithArgs("ABCDEFAB").
		WillReturnRows(rows)

	repo := newUserRepositoryWithSQL(nil, db)
	exists, err := repo.inviteAliasExists(context.Background(), "ABCDEFAB")
	require.NoError(t, err)
	require.True(t, exists)
	require.NoError(t, mock.ExpectationsWereMet())
}

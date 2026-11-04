package timeoptions

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestStoreOptionsAndAdminEdits(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	store := Store{db}
	mock.ExpectQuery("SELECT id,label,seconds,enabled,sort_order").WithArgs(false).WillReturnRows(sqlmock.NewRows([]string{"id", "label", "seconds", "enabled", "sort_order"}).AddRow("10-seconds", "10 seconds", 10, true, 10).AddRow("1-minute", "1 minute", 60, true, 20))
	options, err := store.List(context.Background(), false)
	require.NoError(t, err)
	require.Len(t, options, 2)
	mock.ExpectExec("INSERT INTO question_time_options").WithArgs("10-seconds", "Quick round", 15, false, 30).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, store.Put(context.Background(), Option{"10-seconds", "Quick round", 15, false, 30}))
	require.Error(t, store.Put(context.Background(), Option{"bad|id", "Bad", 1, true, 0}))
	require.NoError(t, mock.ExpectationsWereMet())
}

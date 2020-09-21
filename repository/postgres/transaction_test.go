//+build integration

package postgres

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hmoragrega/paybile/service"
)

func TestListTransactions(t *testing.T) {
	db := setupTestDB(t)
	tt := []struct {
		name     string
		walletID uuid.UUID
		options  service.ListOptions
		wantList service.TransactionList
		wantErr  error
	}{
		{
			name:     "no transactions",
			walletID: walletC,
		},
		{
			name:     "first page",
			walletID: walletD,
			options: service.ListOptions{
				FromID:  nil,
				PerPage: 2,
			},
			wantList: service.TransactionList{
				Results: []service.Transaction{{
					ID:      uuid.MustParse("72db21de-9a63-40c6-b666-d35cf8437fd5"),
					Amount:  09.50,
					Balance: 09.50,
					Type:    service.DepositType,
					Date:    rfc3339MustParse(t, "2020-09-20T10:00:01Z"),
				}, {
					ID:      uuid.MustParse("be3c602a-4410-475b-b39f-016c451726a1"),
					Amount:  08.50,
					Balance: 18.00,
					Type:    service.DepositType,
					Date:    rfc3339MustParse(t, "2020-09-20T10:00:02Z"),
				}},
				NextID: ptrToUUID(uuid.MustParse("ccf28188-92c7-4a30-8f60-70345694f893")),
			},
		},
		{
			name:     "last page",
			walletID: walletD,
			options: service.ListOptions{
				FromID:  ptrToUUID(uuid.MustParse("ccf28188-92c7-4a30-8f60-70345694f893")),
				PerPage: 2,
			},
			wantList: service.TransactionList{
				Results: []service.Transaction{{
					ID:      uuid.MustParse("ccf28188-92c7-4a30-8f60-70345694f893"),
					Amount:  12.5,
					Balance: 30.50,
					Type:    service.DepositType,
					Date:    rfc3339MustParse(t, "2020-09-20T10:00:02Z"),
				}},
			},
		},
		{
			name:     "descending order",
			walletID: walletD,
			options: service.ListOptions{
				Order:   service.Descending,
				PerPage: 2,
			},
			wantList: service.TransactionList{
				Results: []service.Transaction{{
					ID:      uuid.MustParse("ccf28188-92c7-4a30-8f60-70345694f893"),
					Amount:  12.5,
					Balance: 30.50,
					Type:    service.DepositType,
					Date:    rfc3339MustParse(t, "2020-09-20T10:00:02Z"),
				}, {
					ID:      uuid.MustParse("be3c602a-4410-475b-b39f-016c451726a1"),
					Amount:  08.50,
					Balance: 18.00,
					Type:    service.DepositType,
					Date:    rfc3339MustParse(t, "2020-09-20T10:00:02Z"),
				}},
				NextID: ptrToUUID(uuid.MustParse("72db21de-9a63-40c6-b666-d35cf8437fd5")),
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			setupFixtures(ctx, t, db)
			r := TransactionRepository{DB: db}

			got, err := r.ListTransactions(ctx, tc.walletID, tc.options)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("unexpected error: got: %v, want %v", err, tc.wantErr)
			}
			if !reflect.DeepEqual(got, tc.wantList) {
				t.Fatalf("unexpected transaction list: \n got:  %+v \n want: %+v", got, tc.wantList)
			}
		})
	}
}

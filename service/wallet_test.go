package service_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hmoragrega/paybile/service"
	"github.com/hmoragrega/paybile/service/mocks"
)

func TestGetWallet(t *testing.T) {
	t.Parallel()
	var (
		ctx      = context.Background()
		dummyErr = fmt.Errorf("dummy error")
		walletID = uuid.New()
		user     = service.User{ID: uuid.New()}
		wallet   = service.Wallet{ID: walletID, UserID: user.ID, Balance: 20}
	)
	tt := []struct {
		name       string
		user       service.User
		walletID   uuid.UUID
		expect     func(*mocks.WalletReader)
		wantWallet service.Wallet
		wantErr    error
	}{
		{
			name:     "reader error",
			user:     user,
			walletID: walletID,
			expect: func(r *mocks.WalletReader) {
				r.On("GetByID", ctx, walletID).Return(service.Wallet{}, dummyErr)
			},
			wantErr: dummyErr,
		}, {
			name:     "access denied",
			user:     user,
			walletID: walletID,
			expect: func(r *mocks.WalletReader) {
				r.On("GetByID", ctx, walletID).Return(service.Wallet{}, nil)
			},
			wantErr: service.ErrWalletAccessDenied,
		}, {
			name:     "ok",
			user:     user,
			walletID: walletID,
			expect: func(r *mocks.WalletReader) {
				r.On("GetByID", ctx, walletID).Return(wallet, nil)
			},
			wantWallet: wallet,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			wr := &mocks.WalletReader{}
			tc.expect(wr)

			svc := service.WalletService{
				WalletReader: wr,
			}

			got, err := svc.GetWallet(ctx, tc.user, tc.walletID)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("unexpected error: got: %v, want %v", err, tc.wantErr)
			}
			if !reflect.DeepEqual(got, tc.wantWallet) {
				t.Fatalf("unexpected wallet: \n got:  %+v \n want: %+v", got, tc.wantWallet)
			}
		})
	}
}

func TestTransferFunds(t *testing.T) {
	t.Parallel()
	var (
		ctx           = context.Background()
		dummyErr      = fmt.Errorf("dummy error")
		issuerID      = uuid.New()
		originID      = uuid.New()
		destinationID = uuid.New()
		message       = "foo"
		request       = service.TransferRequest{
			Issuer:              service.User{ID: issuerID},
			OriginWalletID:      originID,
			DestinationWalletID: destinationID,
			Amount:              25,
			Message:             &message,
		}
		transfer = service.Transfer{
			ID:                  uuid.New(),
			IssuerID:            issuerID,
			OriginWalletID:      originID,
			DestinationWalletID: destinationID,
			Amount:              25,
			Message:             &message,
			Date:                time.Now(),
		}
	)
	tt := []struct {
		name         string
		req          service.TransferRequest
		walletID     uuid.UUID
		expect       func(*mocks.TransferCreator)
		wantTransfer service.Transfer
		wantErr      error
	}{
		{
			name: "invalid amount error",
			req: service.TransferRequest{
				Amount: 0,
			},
			wantErr: service.ErrInvalidTransactionAmount,
		}, {
			name: "equal wallets",
			req: service.TransferRequest{
				OriginWalletID:      originID,
				DestinationWalletID: originID,
				Amount:              10,
			},
			wantErr: service.ErrSameTransferWallets,
		}, {
			name: "transfer error",
			req:  request,
			expect: func(r *mocks.TransferCreator) {
				r.On("CreateTransfer", ctx, request).Return(transfer, dummyErr)
			},
			wantErr: dummyErr,
		}, {
			name: "transfer ok",
			req:  request,
			expect: func(r *mocks.TransferCreator) {
				r.On("CreateTransfer", ctx, request).Return(transfer, nil)
			},
			wantTransfer: transfer,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := &mocks.TransferCreator{}
			tc.expect(c)

			svc := service.WalletService{
				TransferCreator: c,
			}

			got, err := svc.TransferFunds(ctx, tc.req)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("unexpected error: got: %v, want %v", err, tc.wantErr)
			}
			if !reflect.DeepEqual(got, tc.wantTransfer) {
				t.Fatalf("unexpected transfer: \n got:  %+v \n want: %+v", got, tc.wantTransfer)
			}
		})
	}
}

func TestListTransactions(t *testing.T) {
	t.Parallel()
	var (
		ctx      = context.Background()
		dummyErr = fmt.Errorf("dummy error")
		issuerID = uuid.New()
		user     = service.User{ID: issuerID}
		walletID = uuid.New()
		wallet   = service.Wallet{ID: walletID, UserID: issuerID}
		opts     = service.ListOptions{PerPage: 5, FromID: &issuerID}
		list     = service.TransactionList{Results: []service.Transaction{{ID: uuid.New()}}}
	)
	tt := []struct {
		name     string
		user     service.User
		walletID uuid.UUID
		opts     service.ListOptions
		expect   func(*mocks.WalletReader, *mocks.TransactionsReader)
		wantList service.TransactionList
		wantErr  error
	}{
		{
			name:     "error loading wallet",
			walletID: walletID,
			expect: func(wr *mocks.WalletReader, tr *mocks.TransactionsReader) {
				wr.On("GetByID", ctx, walletID).Return(service.Wallet{}, dummyErr)
			},
			wantErr: dummyErr,
		}, {
			name:     "wallet access denied",
			walletID: walletID,
			user:     user,
			expect: func(wr *mocks.WalletReader, tr *mocks.TransactionsReader) {
				wr.On("GetByID", ctx, walletID).Return(service.Wallet{}, nil)
			},
			wantErr: service.ErrWalletAccessDenied,
		}, {
			name:     "list error",
			walletID: walletID,
			user:     user,
			opts:     opts,
			expect: func(wr *mocks.WalletReader, tr *mocks.TransactionsReader) {
				wr.On("GetByID", ctx, walletID).Return(wallet, nil)
				tr.On("ListTransactions", ctx, walletID, opts).Return(service.TransactionList{}, dummyErr)
			},
			wantErr: dummyErr,
		}, {
			name:     "list ok",
			walletID: walletID,
			user:     user,
			opts:     opts,
			expect: func(wr *mocks.WalletReader, tr *mocks.TransactionsReader) {
				wr.On("GetByID", ctx, walletID).Return(wallet, nil)
				tr.On("ListTransactions", ctx, walletID, opts).Return(list, nil)
			},
			wantList: list,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			wr := &mocks.WalletReader{}
			tr := &mocks.TransactionsReader{}
			tc.expect(wr, tr)

			svc := service.WalletService{
				WalletReader:       wr,
				TransactionsReader: tr,
			}

			got, err := svc.ListTransactions(ctx, tc.user, tc.walletID, tc.opts)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("unexpected error: got: %v, want %v", err, tc.wantErr)
			}
			if err == nil && !reflect.DeepEqual(got, tc.wantList) {
				t.Fatalf("unexpected list: \n got:  %+v \n want: %+v", got, tc.wantList)
			}
		})
	}
}

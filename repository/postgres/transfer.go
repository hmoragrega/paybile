package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/hmoragrega/paybile/service"
)

var (
	// Specific postgres implementation errors.
	errTxBegin  = errors.New("cannot begin transaction")
	errTxCommit = errors.New("cannot commit transaction")
)

type TransferRepository struct {
	DB              *sql.DB
	TransactionRepo *TransactionRepository
	WalletRepo      *WalletRepository
}

type queryHandler interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

func (r *TransferRepository) CreateTransfer(ctx context.Context, req service.TransferRequest) (t service.Transfer, err error) {
	if req.Amount <= 0 {
		return t, service.ErrInvalidTransactionAmount
	}

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return t, fmt.Errorf("%w: %v", errTxBegin, err)
	}
	defer func() {
		if err != nil {
			err = rollback(tx, err)
		}
	}()

	wallets, err := r.WalletRepo.findByID(ctx, tx, req.OriginWalletID, req.DestinationWalletID)
	if err != nil {
		return t, err
	}
	origin, ok := wallets[req.OriginWalletID]
	if !ok {
		return t, fmt.Errorf("%w: origin wallet not found", service.ErrWalletNotFound)
	}
	destination, ok := wallets[req.DestinationWalletID]
	if !ok {
		return t, fmt.Errorf("%w: destination wallet not found", service.ErrWalletNotFound)
	}

	if !req.Issuer.CanWrite(origin) {
		return t, service.ErrWalletAccessDenied
	}
	if w := origin; w.Balance < req.Amount {
		return t, fmt.Errorf("%w: balance %.2f", service.ErrInsufficientFunds, w.Balance)
	}

	row := tx.QueryRowContext(ctx, `
		INSERT INTO transfers (issuer_id, origin_wallet_id, destination_wallet_id, amount, message)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, date`,
		req.Issuer.ID, req.OriginWalletID, req.DestinationWalletID, req.Amount, req.Message,
	)
	if err = row.Scan(&t.ID, &t.Date); err != nil {
		return t, fmt.Errorf("cannot insert transfer: %v", err)
	}

	// Insert transactions for origin and destination with opposite amounts.
	for _, x := range []struct {
		wallet *service.Wallet
		amount float64
	}{{
		wallet: &origin,
		amount: req.Amount * -1,
	}, {
		wallet: &destination,
		amount: req.Amount,
	}} {
		balance := x.wallet.Balance + x.amount
		_, err = r.TransactionRepo.insert(ctx, tx, x.wallet.ID, x.amount, balance, service.TransferType, &t.ID)
		if err != nil {
			return t, err
		}
		if err = r.WalletRepo.updateBalance(ctx, tx, x.wallet.ID, balance); err != nil {
			return t, err
		}
	}

	if err = tx.Commit(); err != nil {
		return t, fmt.Errorf("%w: %v", errTxCommit, err)
	}

	t.IssuerID = req.Issuer.ID
	t.OriginWalletID = req.OriginWalletID
	t.DestinationWalletID = req.DestinationWalletID
	t.Amount = req.Amount
	t.Message = req.Message

	return t, nil
}

func rollback(tx *sql.Tx, err error) error {
	if rollbackErr := tx.Rollback(); rollbackErr != nil {
		err = fmt.Errorf("rollback trigger %w: rollback error %v", err, rollbackErr)
	}

	return err
}

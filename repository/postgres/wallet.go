package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/hmoragrega/paybile/service"
	"github.com/lib/pq"
)

type WalletRepository struct {
	DB *sql.DB
}

func (r *WalletRepository) GetByID(ctx context.Context, walletID uuid.UUID) (w service.Wallet, err error) {
	query := `SELECT id, user_id, balance FROM wallets WHERE id = $1`
	row := r.DB.QueryRowContext(ctx, query, walletID)
	err = row.Scan(&w.ID, &w.UserID, &w.Balance)
	if err == sql.ErrNoRows {
		return w, service.ErrWalletNotFound
	}
	if err != nil {
		return w, fmt.Errorf("%w: %v", fmt.Errorf("cannot retrieve wallet"), err)
	}

	return w, nil
}

func (r *WalletRepository) updateBalance(
	ctx context.Context,
	db queryHandler,
	walletID uuid.UUID,
	balance float64,
) error {
	_, err := db.ExecContext(ctx, `
		UPDATE wallets
		SET balance = $1
		WHERE id = $2`,
		balance, walletID,
	)
	if err != nil {
		return fmt.Errorf("cannot update wallet balance: %v", err)
	}

	return nil
}

func (r *WalletRepository) findByID(
	ctx context.Context,
	db queryHandler,
	walletIDs ...uuid.UUID,
) (res map[uuid.UUID]service.Wallet, err error) {
	ids := make([]interface{}, len(walletIDs))
	for i, x := range walletIDs {
		ids[i] = x
	}

	query := `SELECT id, user_id, balance FROM wallets WHERE id = ANY($1)`
	rows, err := db.QueryContext(ctx, query, pq.Array(ids))
	if err != nil {
		return res, fmt.Errorf("%w: %v", fmt.Errorf("cannot retrieve wallets"), err)
	}

	res = make(map[uuid.UUID]service.Wallet)
	for rows.Next() {
		var w service.Wallet
		if err := rows.Scan(&w.ID, &w.UserID, &w.Balance); err != nil {
			return res, fmt.Errorf("%w: %v", fmt.Errorf("cannot scan wallet"), err)
		}
		res[w.ID] = w
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %v", fmt.Errorf("cannot iterate wallets"), err)
	}

	return res, nil
}

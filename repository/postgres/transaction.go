package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/hmoragrega/paybile/service"
)

type TransactionRepository struct {
	DB *sql.DB
}

func (r *TransactionRepository) insert(
	ctx context.Context,
	db queryHandler,
	walletID uuid.UUID,
	amount float64,
	balance float64,
	transactionType service.TransactionType,
	referenceID *uuid.UUID,
) (t service.Transaction, err error) {
	row := db.QueryRowContext(ctx, `
		INSERT INTO transactions(wallet_id, amount, balance, transaction_type, reference_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, date`,
		walletID, amount, balance, transactionType, referenceID,
	)

	err = row.Scan(&t.ID, &t.Date)
	if err != nil {
		return t, fmt.Errorf("cannot insert transaction: %v", err)
	}

	t.Amount = amount
	t.Balance = balance
	t.Type = transactionType
	t.ReferenceID = referenceID

	return t, nil
}

func (r *TransactionRepository) ListTransactions(
	ctx context.Context,
	walletID uuid.UUID,
	opt service.ListOptions,
) (res service.TransactionList, err error) {

	order := "ASC"
	if opt.Order.Descending() {
		order = "DESC"
	}

	// Request one more elements to get the next ID if available.
	params := []interface{}{walletID, opt.PerPage + 1}

	var where string
	if opt.FromID != nil {
		if opt.Order.Ascending() {
			where = ` AND t.ID >= $3 `
		} else {
			where = ` AND t.ID <= $3 `
		}
		params = append(params, *opt.FromID)
	}

	query := fmt.Sprintf(`
		SELECT t.id, t.amount, t.date, t.balance, t.transaction_type, t.reference_id
		FROM transactions t
		WHERE t.wallet_id = $1 %s
		ORDER BY t.date %s, t.id %s
		LIMIT $2`,
		where, order, order,
	)

	rows, err := r.DB.QueryContext(ctx, query, params...)
	if err != nil {
		return res, fmt.Errorf("cannot query transactions: %v", err)
	}

	for rows.Next() {
		var tx service.Transaction
		if err := rows.Scan(&tx.ID, &tx.Amount, &tx.Date, &tx.Balance, &tx.Type, &tx.ReferenceID); err != nil {
			return res, fmt.Errorf("%w: %v", fmt.Errorf("cannot scan transaction"), err)
		}
		tx.Date = tx.Date.UTC()
		res.Results = append(res.Results, tx)
	}
	if err = rows.Err(); err != nil {
		return res, fmt.Errorf("%w: %v", fmt.Errorf("cannot iterate transactions"), err)
	}

	// Return a cursor to the next transaction ID if there are more results.
	if res.Results != nil && len(res.Results) > opt.PerPage {
		res.NextID = &res.Results[opt.PerPage].ID
		res.Results = res.Results[:opt.PerPage]
	}

	return res, nil
}

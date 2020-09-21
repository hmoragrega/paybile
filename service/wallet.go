package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var (
	ErrWalletAccessDenied       = errors.New("wallet access denied")
	ErrInvalidTransactionAmount = errors.New("invalid transaction amount")
	ErrWalletNotFound           = errors.New("wallet not found")
	ErrInsufficientFunds        = errors.New("origin has insufficient funds")
	ErrSameTransferWallets      = errors.New("same origin and destination")
)

type WalletReader interface {
	// GetByID returns the wallet by the ID.
	// Errors:
	// - ErrWalletNotFound: if the wallet doesn not exists.
	GetByID(ctx context.Context, walletID uuid.UUID) (Wallet, error)
}

type TransferCreator interface {
	// TransferFunds atomically transfer funds between two wallets.
	// Errors:
	// - ErrInvalidTransactionAmount: if the amount is zero or less.
	// - ErrWalletNotFound: if a wallet involved does not exists.
	// - ErrWalletAccessDenied: if the issuer cannot transfer funds from the wallet.
	// - ErrInsufficientFunds: if the origin wallet does not have enough funds.
	CreateTransfer(ctx context.Context, req TransferRequest) (Transfer, error)
}

type TransactionsReader interface {
	// ListTransactions returns a list transactions of a wallet ordered by date.
	// Returns:
	// - list: transactions for the current page.
	// - next: the next transaction ID in the list, can be used to request the next page.
	ListTransactions(ctx context.Context, walletID uuid.UUID, filters ListOptions) (list TransactionList, err error)
}

type WalletService struct {
	WalletReader       WalletReader
	TransferCreator    TransferCreator
	TransactionsReader TransactionsReader
}

type TransferRequest struct {
	Issuer              User
	OriginWalletID      uuid.UUID
	DestinationWalletID uuid.UUID
	Amount              float64
	Message             *string
}

type TransactionList struct {
	Results []Transaction `json:"results"`
	NextID  *uuid.UUID    `json:"next_id"`
}

func (svc *WalletService) GetWallet(ctx context.Context, user User, walletID uuid.UUID) (w Wallet, err error) {
	w, err = svc.WalletReader.GetByID(ctx, walletID)
	if err != nil {
		return w, fmt.Errorf("cannot get wallet: %w", err)
	}
	if !user.CanRead(w) {
		return w, ErrWalletAccessDenied
	}

	return w, nil
}

// Transfer transfer funds to another wallet.
func (svc *WalletService) TransferFunds(ctx context.Context, req TransferRequest) (t Transfer, err error) {
	if req.Amount <= 0 {
		return t, ErrInvalidTransactionAmount
	}
	if req.DestinationWalletID == req.OriginWalletID {
		return t, ErrSameTransferWallets
	}

	t, err = svc.TransferCreator.CreateTransfer(ctx, req)
	if err != nil {
		return t, fmt.Errorf("cannot transfer funds: %w", err)
	}

	return t, nil
}

func (svc *WalletService) ListTransactions(
	ctx context.Context,
	user User,
	walletID uuid.UUID,
	opt ListOptions,
) (l TransactionList, err error) {
	w, err := svc.WalletReader.GetByID(ctx, walletID)
	if err != nil {
		return l, fmt.Errorf("cannot get wallet: %w", err)
	}
	if !user.CanRead(w) {
		return l, ErrWalletAccessDenied
	}

	if opt.PerPage <= 0 {
		opt.PerPage = DefaultPerPage
	}

	l, err = svc.TransactionsReader.ListTransactions(ctx, walletID, opt)
	if err != nil {
		return l, fmt.Errorf("%w: cannot list transactions", err)
	}

	return l, nil
}

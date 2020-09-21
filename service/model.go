package service

import (
	"time"

	"github.com/google/uuid"
)

type UserResource interface {
	// IsOwner checks if the user is the owner of the resource.
	IsOwner(user User) bool
}

type User struct {
	ID             uuid.UUID
	Login          string
	HashedPassword string
}

// CanRead checks if the user has read access to the resource.
func (u User) CanRead(r UserResource) bool {
	return r.IsOwner(u) // TODO or is admin?
}

// CanWrite checks if the user has write access to the resource.
func (u User) CanWrite(r UserResource) bool {
	return r.IsOwner(u) // TODO or is admin?
}

// IsCorrectPass checks whether the given password is correct.
func (u User) IsCorrectPass(password string) bool {
	return u.HashedPassword == password
}

type Wallet struct {
	ID      uuid.UUID `json:"id"`
	UserID  uuid.UUID `json:"user_id"`
	Balance float64   `json:"balance"`
}

// IsOwner checks if the user is the owner of the wallet.
func (w Wallet) IsOwner(u User) bool {
	return w.UserID == u.ID
}

type TransactionType string

const (
	TransferType TransactionType = "transfer"
	DepositType                  = "deposit"
)

type Transaction struct {
	ID          uuid.UUID       `json:"id"`
	Amount      float64         `json:"amount"`
	Balance     float64         `json:"balance"`
	Type        TransactionType `json:"transaction_type"`
	ReferenceID *uuid.UUID      `json:"reference_id"`
	Date        time.Time       `json:"date"`
}

type Transfer struct {
	ID                  uuid.UUID `json:"id"`
	IssuerID            uuid.UUID `json:"issuer_id"`
	OriginWalletID      uuid.UUID `json:"origin_wallet_id"`
	DestinationWalletID uuid.UUID `json:"destination_wallet_id"`
	Amount              float64   `json:"amount"`
	Message             *string   `json:"message"`
	Date                time.Time `json:"date"`
}

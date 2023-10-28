// Description: Account package models and errors.

package account

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

const (
	AccountIdPrefix = "account-"
)

var (
	ErrInvalidAmount     = errors.New("invalid amount")
	ErrInsufficientFunds = errors.New("insufficient funds")
)

type Account struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Balance float64   `json:"balance,string"`
}

func NewAccount(name string, balance float64) *Account {
	return &Account{
		ID:      uuid.New(),
		Name:    name,
		Balance: balance,
	}
}

func (a *Account) GetID() string {
	return fmt.Sprintf("%s-%s", AccountIdPrefix, a.ID.String())
}

type TransferRequest struct {
	Sender   string  `json:"sender"`
	Reciever string  `json:"reciever"`
	Amount   float64 `json:"amount"`
}

// ValidateAmount validates the transfer request amount against the sender's balance.
// Returns an error if the amount is invalid or insufficient.
func (t TransferRequest) ValidateAmount(sender *Account) error {
	if t.Amount <= 0 {
		return ErrInvalidAmount
	}

	if t.Amount > sender.Balance {
		return ErrInsufficientFunds
	}

	return nil
}

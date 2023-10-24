package account

import (
	"fmt"

	"github.com/google/uuid"
)

const (
	AccountIdPrefix = "account-"
)

type Account struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Balance float64   `json:"balance,string"`
}

func (a *Account) GetID() string {
	return fmt.Sprintf("%s-%s", AccountIdPrefix, a.ID.String())
}

package account

import (
	"github.com/google/uuid"
)

type Account struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Balance float64   `json:"balance,string"`
}

func (a *Account) GetID() string {
	return a.ID.String()
}

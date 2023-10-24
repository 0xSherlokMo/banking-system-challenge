package account

import (
	"sync"

	"github.com/google/uuid"
)

type Account struct {
	ID      uuid.UUID  `json:"id"`
	Name    string     `json:"name"`
	Balance float64    `json:"balance"`
	latch   sync.Mutex `json:"-"`
}

func (a *Account) GetID() string {
	return a.ID.String()
}

func (a *Account) Lock() {
	a.latch.Lock()
}

func (a *Account) Unlock() {
	a.latch.Unlock()
}

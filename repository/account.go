package repository

import (
	"github.com/0xSherlokMo/banking-system-challenge/account"
	"github.com/0xSherlokMo/banking-system-challenge/calculator"
	"github.com/0xSherlokMo/banking-system-challenge/ctx"
	"github.com/0xSherlokMo/banking-system-challenge/memorydb"
)

type AccountRepository struct {
	ctx *ctx.DefaultContext
}

func NewAccountRepository(ctx *ctx.DefaultContext) *AccountRepository {
	return &AccountRepository{
		ctx: ctx,
	}
}

type LockReleaser = func()

func (a *AccountRepository) All(safe bool) []*account.Account {
	database := a.ctx.MemoryDB()
	accounts := database.GetM(database.Keys(), memorydb.Opts{
		Safe: safe,
	})
	return accounts
}

func (a *AccountRepository) GetByKey(key memorydb.Key, safe bool) (*account.Account, error) {
	database := a.ctx.MemoryDB()
	account, err := database.Get(key, memorydb.Opts{
		Safe: safe,
	})
	if err != nil {
		return nil, err
	}
	return account, nil
}

// !!! This method is not thread safe !!!
// You should use PrepareAccounts, Commit and Rollback methods to make it thread safe.
func (a *AccountRepository) TransferMoney(request account.TransferRequest) (*account.Account, error) {
	database := a.ctx.MemoryDB()
	senderAccount, err := database.Get(request.Sender, memorydb.Opts{
		Safe: memorydb.ConcurrentNotSafe,
	})
	if err != nil {
		a.ctx.Logger().Errorw("account locked but doesn't exist", "account", request.Sender)
		return nil, err
	}
	receiverAccount, err := database.Get(request.Reciever, memorydb.Opts{
		Safe: memorydb.ConcurrentNotSafe,
	})
	if err != nil {
		a.ctx.Logger().Errorw("account locked but doesn't exist", "account", request.Reciever)
		return nil, err
	}

	if err := request.ValidateAmount(senderAccount); err != nil {
		a.ctx.Logger().Debugw("invalid amount", "request", request, "error", err)
		return nil, err
	}

	senderAccount.Balance = calculator.PreciseAdd(senderAccount.Balance, -request.Amount)
	database.Set(request.Sender, senderAccount, memorydb.Opts{Safe: memorydb.ConcurrentNotSafe})
	receiverAccount.Balance = calculator.PreciseAdd(receiverAccount.Balance, request.Amount)
	database.Set(request.Reciever, receiverAccount, memorydb.Opts{Safe: memorydb.ConcurrentNotSafe})

	return senderAccount, nil
}

func (a *AccountRepository) PrepareAccounts(keys ...memorydb.Key) error {
	database := a.ctx.MemoryDB()
	var releasers []LockReleaser
	for _, key := range keys {
		err := database.Lock(key)
		if err != nil {
			a.ctx.Logger().Errorw("Cannot lock account", "account", key, "error", err)
			go func() {
				for _, releaser := range releasers {
					releaser()
				}
			}()
			return err
		}

		releasers = append(releasers, func() {
			a.Rollback(key)
		})
	}

	return nil
}

// added for readability
func (a *AccountRepository) Rollback(keys ...memorydb.Key) {
	a.Rollback(keys...)
}

// added for readability
func (a *AccountRepository) Commit(keys ...memorydb.Key) {
	a.unlock(keys...)
}

func (a *AccountRepository) unlock(keys ...memorydb.Key) {
	database := a.ctx.MemoryDB()
	for _, key := range keys {
		err := database.Unlock(key)
		if err != nil {
			a.ctx.Logger().Errorw("Cannot unlock account", "account", key, "error", err)
		}
	}
}

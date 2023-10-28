package repository_test

import (
	"sync"
	"testing"
	"time"

	"github.com/0xSherlokMo/banking-system-challenge/account"
	"github.com/0xSherlokMo/banking-system-challenge/ctx"
	"github.com/0xSherlokMo/banking-system-challenge/memorydb"
	"github.com/0xSherlokMo/banking-system-challenge/repository"
)

type TransferTestDirection int

const (
	FromFirstToSecond TransferTestDirection = iota
	FromSecondToFirst
)

type MoneyTransferOperation struct {
	Direction TransferTestDirection
	Amount    float64
}

func NewMoneyTransferOperation(direction TransferTestDirection, amount float64) MoneyTransferOperation {
	return MoneyTransferOperation{
		Direction: direction,
		Amount:    amount,
	}
}

type MoneyTransferTest struct {
	FirstAccount    *account.Account
	SecondAccount   *account.Account
	operations      []MoneyTransferOperation
	expectedAmounts []float64
}

func TestMoneyTransfer(t *testing.T) {
	ctx, tt := LoadMoneyTransferTestTable()
	repositoryMock := repository.NewAccountRepository(&ctx)
	for id, tc := range tt {
		var wg sync.WaitGroup
		for _, o := range tc.operations {
			wg.Add(1)
			go func(operation MoneyTransferOperation) {
				defer wg.Done()
				backoffInMillis := 10
				// should add circuit breaker here but skipped for simplicity
				for {
					err := repositoryMock.PrepareAccounts(tc.FirstAccount.GetID(), tc.SecondAccount.GetID())
					if err != nil {
						time.Sleep(time.Duration(backoffInMillis) * time.Millisecond)
						backoffInMillis *= 2
						continue
					}
					request := account.TransferRequest{
						Amount: operation.Amount,
					}
					switch operation.Direction {
					case FromFirstToSecond:
						request.Sender = tc.FirstAccount.GetID()
						request.Reciever = tc.SecondAccount.GetID()
					case FromSecondToFirst:
						request.Sender = tc.SecondAccount.GetID()
						request.Reciever = tc.FirstAccount.GetID()
					}
					_, err = repositoryMock.TransferMoney(request)
					if err != nil {
						repositoryMock.Rollback(tc.FirstAccount.GetID(), tc.SecondAccount.GetID())
						time.Sleep(time.Duration(backoffInMillis) * time.Millisecond)
						backoffInMillis *= 2
						continue
					}
					repositoryMock.Commit(tc.FirstAccount.GetID(), tc.SecondAccount.GetID())
					break
				}
			}(o)
		}
		wg.Wait()
		firstAccount, _ := repositoryMock.GetByKey(tc.FirstAccount.GetID(), memorydb.ConcurrentSafe)
		secondAccount, _ := repositoryMock.GetByKey(tc.SecondAccount.GetID(), memorydb.ConcurrentSafe)
		if firstAccount.Balance != tc.expectedAmounts[0] {
			t.Errorf("expected first account to have %f but got %f, account %+v, tc %d", tc.expectedAmounts[0], firstAccount.Balance, firstAccount, id)
		}
		if secondAccount.Balance != tc.expectedAmounts[1] {
			t.Errorf("expected second account to have %f but got %f, account %+v, tc %d", tc.expectedAmounts[1], secondAccount.Balance, secondAccount, id)
		}
	}
}

func LoadMoneyTransferTestTable() (ctx.DefaultContext, []MoneyTransferTest) {
	ctx := ctx.NewDefaultContext().WithMemoryDB()

	testTable := []MoneyTransferTest{
		{
			FirstAccount:  account.NewAccount("mario", 100),
			SecondAccount: account.NewAccount("jack", 0),
			operations: []MoneyTransferOperation{
				NewMoneyTransferOperation(FromFirstToSecond, 50),
				NewMoneyTransferOperation(FromSecondToFirst, 50),
				NewMoneyTransferOperation(FromFirstToSecond, 10),
				NewMoneyTransferOperation(FromFirstToSecond, 10),
				NewMoneyTransferOperation(FromSecondToFirst, 20),
			},
			expectedAmounts: []float64{
				100,
				0,
			},
		},
		{
			FirstAccount:  account.NewAccount("hello-kitty", 100),
			SecondAccount: account.NewAccount("super-mario", 0),
			operations: []MoneyTransferOperation{
				NewMoneyTransferOperation(FromFirstToSecond, 20),
				NewMoneyTransferOperation(FromFirstToSecond, 5),
				NewMoneyTransferOperation(FromSecondToFirst, 25),
			},
			expectedAmounts: []float64{
				100,
				0,
			},
		},
		{
			// evil guy trying to double his money by transferring money to his friend
			FirstAccount:  account.NewAccount("evil-guy", 100),
			SecondAccount: account.NewAccount("friend-to-evil-guy", 100),
			operations: []MoneyTransferOperation{
				NewMoneyTransferOperation(FromFirstToSecond, 100),
				NewMoneyTransferOperation(FromSecondToFirst, 100),
				NewMoneyTransferOperation(FromFirstToSecond, 100),
				NewMoneyTransferOperation(FromSecondToFirst, 100),
				NewMoneyTransferOperation(FromFirstToSecond, 100),
				NewMoneyTransferOperation(FromSecondToFirst, 100),
				NewMoneyTransferOperation(FromFirstToSecond, 100),
				NewMoneyTransferOperation(FromSecondToFirst, 100),
			},
			expectedAmounts: []float64{
				100,
				100,
			},
		},
		{
			// lucky man trying to make poor guy have negative balance
			FirstAccount:  account.NewAccount("lucky-guy", 100),
			SecondAccount: account.NewAccount("poor-guy", 1),
			operations: []MoneyTransferOperation{
				NewMoneyTransferOperation(FromSecondToFirst, 99),
				NewMoneyTransferOperation(FromFirstToSecond, 99),
				NewMoneyTransferOperation(FromSecondToFirst, 1),
				NewMoneyTransferOperation(FromFirstToSecond, 1),
				NewMoneyTransferOperation(FromSecondToFirst, 1),
			},
			expectedAmounts: []float64{
				101,
				0,
			},
		},
	}

	db := ctx.MemoryDB()
	for _, tc := range testTable {
		db.Setnx(tc.FirstAccount.GetID(), tc.FirstAccount)
		db.Setnx(tc.SecondAccount.GetID(), tc.SecondAccount)
	}

	return *ctx, testTable
}

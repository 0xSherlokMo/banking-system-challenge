package router

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/0xSherlokMo/banking-system-challenge/account"
	"github.com/0xSherlokMo/banking-system-challenge/calculator"
	"github.com/0xSherlokMo/banking-system-challenge/ctx"
	"github.com/0xSherlokMo/banking-system-challenge/memorydb"
	"github.com/gin-gonic/gin"
)

type AccountRouter struct {
	ctx *ctx.DefaultContext
}

func InstallAccountRouter(engine *gin.Engine, ctx *ctx.DefaultContext) AccountRouter {
	accountRouter := AccountRouter{
		ctx: ctx,
	}

	accountRouter.install(
		engine.Group("/accounts"),
	)

	return accountRouter
}

func (a *AccountRouter) install(router *gin.RouterGroup) {
	router.GET("/", a.getAll)
	router.GET("/:id", a.getId)
	router.POST("/:from/transfer/:to", a.transfer)
}

func (a *AccountRouter) getAll(c *gin.Context) {
	repository := a.ctx.MemoryDB()

	safe := c.Query("safe") == "true"

	accounts := repository.GetM(repository.Keys(), memorydb.Opts{
		Safe: safe,
	})

	c.JSON(http.StatusOK, gin.H{
		"accounts": accounts,
	})
}

func (a *AccountRouter) getId(c *gin.Context) {
	repository := a.ctx.MemoryDB()

	key := fmt.Sprintf("%s-%s", account.AccountIdPrefix, c.Param("id"))
	account, err := repository.Get(key, memorydb.Opts{
		Safe: true,
	})

	if errors.Is(err, memorydb.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "account does not exist",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"account": account,
	})
}

func (a *AccountRouter) transfer(c *gin.Context) {
	repository := a.ctx.MemoryDB()
	var request account.TransferRequest
	err := c.BindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request"})
		return
	}

	request.Sender = fmt.Sprintf("%s-%s", account.AccountIdPrefix, c.Param("from"))
	request.Reciever = fmt.Sprintf("%s-%s", account.AccountIdPrefix, c.Param("to"))
	err = repository.Lock(request.Sender)
	if err != nil {
		status := http.StatusBadRequest
		message := "Sender account does not exist"
		if errors.Is(err, memorydb.ErrRowLocked) {
			a.ctx.Logger().Debugw("account locked", "account", request.Sender)
			status = http.StatusLocked
			message = "Something could've gone wrong. Congrats, you're a survivor."
		}
		c.JSON(status, gin.H{"message": message})
		return
	}
	defer repository.Unlock(request.Sender)

	err = repository.Lock(request.Reciever)
	if err != nil {
		status := http.StatusBadRequest
		message := "Sender account does not exist"
		if errors.Is(err, memorydb.ErrRowLocked) {
			status = http.StatusLocked
			a.ctx.Logger().Debugw("account locked", "account", request.Reciever)
			message = "Something could've gone wrong. Congrats, you're a survivor."
		}
		c.JSON(status, gin.H{"message": message})
		return
	}
	defer repository.Unlock(request.Reciever)

	senderAccount, err := repository.Get(request.Sender, memorydb.Opts{Safe: false})
	if err != nil {
		a.ctx.Logger().Error("account locked but doesn't exist", "account", request.Sender)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong."})
		return
	}
	receiverAccount, err := repository.Get(request.Reciever, memorydb.Opts{Safe: false})
	if err != nil {
		a.ctx.Logger().Error("account locked but doesn't exist", "account", request.Reciever)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong."})
		return
	}

	if err := request.ValidateAmount(senderAccount); err != nil {
		a.ctx.Logger().Debugw("invalid amount", "request", request, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	senderAccount.Balance = calculator.PreciseAdd(senderAccount.Balance, -request.Amount)
	receiverAccount.Balance = calculator.PreciseAdd(receiverAccount.Balance, request.Amount)

	repository.Set(request.Sender, senderAccount, memorydb.Opts{Safe: false})
	repository.Set(request.Reciever, receiverAccount, memorydb.Opts{Safe: false})
	c.JSON(http.StatusOK, gin.H{
		"Balance": senderAccount.Balance,
	})
}

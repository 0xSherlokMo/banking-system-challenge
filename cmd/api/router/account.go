package router

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/0xSherlokMo/banking-system-challenge/account"
	"github.com/0xSherlokMo/banking-system-challenge/ctx"
	"github.com/0xSherlokMo/banking-system-challenge/memorydb"
	"github.com/0xSherlokMo/banking-system-challenge/repository"
	"github.com/gin-gonic/gin"
)

type AccountRouter struct {
	ctx               *ctx.DefaultContext
	AccountRepository *repository.AccountRepository
}

func InstallAccountRouter(engine *gin.Engine, ctx *ctx.DefaultContext) AccountRouter {
	accountRouter := AccountRouter{
		ctx:               ctx,
		AccountRepository: repository.NewAccountRepository(ctx),
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
	var safe bool
	if c.Query("safe") == "true" {
		safe = memorydb.ConcurrentSafe
	}

	accounts := a.AccountRepository.All(safe)

	c.JSON(http.StatusOK, gin.H{
		"accounts": accounts,
	})
}

func (a *AccountRouter) getId(c *gin.Context) {
	key := fmt.Sprintf("%s-%s", account.AccountIdPrefix, c.Param("id"))

	account, err := a.AccountRepository.GetByKey(key, memorydb.ConcurrentSafe)
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
	var request account.TransferRequest
	err := c.BindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request"})
		return
	}
	request.Sender = fmt.Sprintf("%s-%s", account.AccountIdPrefix, c.Param("from"))
	request.Reciever = fmt.Sprintf("%s-%s", account.AccountIdPrefix, c.Param("to"))

	err = a.AccountRepository.PrepareAccounts(request.Sender, request.Reciever)
	if err != nil {
		if errors.Is(err, memorydb.ErrRowLocked) {
			c.JSON(http.StatusLocked, gin.H{"message": "Something could've gone wrong. Congrats, you're a survivor"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"message": "Sender account does not exist"})
		return
	}
	defer a.AccountRepository.Commit(request.Sender, request.Reciever)

	senderAccount, err := a.AccountRepository.TransferMoney(request)
	if err != nil {
		if errors.Is(err, memorydb.ErrRecordExists) {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong."})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Balance": senderAccount.Balance,
	})
}

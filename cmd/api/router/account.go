package router

import (
	"net/http"

	"github.com/0xSherlokMo/banking-system-challenge/ctx"
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
}

func (a *AccountRouter) getAll(c *gin.Context) {
	repository := a.ctx.MemoryDB()

	accounts := repository.GetM(repository.Keys())

	c.JSON(http.StatusOK, gin.H{
		"accounts": accounts,
	})
}

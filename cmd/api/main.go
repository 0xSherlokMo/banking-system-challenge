package main

import (
	"os"

	"github.com/0xSherlokMo/banking-system-challenge/cmd/api/router"
	"github.com/0xSherlokMo/banking-system-challenge/ctx"
	"github.com/gin-gonic/gin"
)

func main() {
	app := ctx.NewDefaultContext().WithMemoryDB().LoadAccounts()
	defer app.Exit()

	engine := gin.Default()
	router.InstallHealthRouter(engine)
	router.InstallAccountRouter(engine, app)
	engine.Run(":" + os.Getenv("PORT"))
}

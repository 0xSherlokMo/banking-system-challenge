package main

import (
	"os"

	"github.com/0xSherlokMo/banking-system-challenge/cmd/api/router"
	"github.com/gin-gonic/gin"
)

func main() {
	engine := gin.Default()
	router.InstallHealthRouter(engine)
	engine.Run(":" + os.Getenv("PORT"))
}

package main

import (
	"github.com/gin-gonic/gin"
	"web/routers"
)

func main() {
	r := gin.Default()
	routers.MemberRouterInit(r)
	routers.LoginRouterInit(r)
	r.Run(":4396")
}

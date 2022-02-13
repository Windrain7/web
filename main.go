package main

import (
	"github.com/gin-gonic/gin"
	"web/routers"
)

func main() {
	r := gin.Default()
	routers.MemberRouterInit(r)
	routers.LoginRouterInit(r)
	routers.CourseRouterInit(r)
	routers.StudentRouterInit(r)
	r.Run(":4396")
}

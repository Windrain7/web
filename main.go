package main

import (
	"encoding/gob"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"web/models"
	"web/routers"
)

func main() {
	gob.Register(models.Member{})
	r := gin.Default()
	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("mysession", store))

	routers.MemberRouterInit(r)
	routers.LoginRouterInit(r)
	routers.CourseRouterInit(r)
	routers.StudentRouterInit(r)
	r.Run(":80")
}

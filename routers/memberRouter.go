package routers

import (
	"github.com/gin-gonic/gin"
	"web/controllers"
)

func MemberRouterInit(c *gin.Engine) {
	g := c.Group("/api/v1/member")
	g.GET("/", controllers.MemberController{}.Index)
	g.POST("/create", controllers.MemberController{}.Create)
	g.GET("/list")
	g.POST("/update")
	g.POST("/delete")
}

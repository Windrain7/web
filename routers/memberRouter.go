package routers

import "github.com/gin-gonic/gin"

func MemberRouterInit(c *gin.Engine) {
	g := c.Group("/api/v1/member")
	g.GET("/")
	g.POST("/create")
	g.GET("/list")
	g.POST("/update")
	g.POST("/delete")
}

package routers

import "github.com/gin-gonic/gin"

func LoginRouterInit(c *gin.Engine) {
	g := c.Group("/api/v1/auth")
	g.GET("/")
	g.POST("/login")
	g.GET("/logout")
	g.POST("/whoami")
}

package routers

import "github.com/gin-gonic/gin"
import "web/controllers"

func LoginRouterInit(c *gin.Engine) {
	g := c.Group("/api/v1/auth")
	g.GET("/")
	g.POST("/login", controllers.LoginController{}.Login)
	g.POST("/logout", controllers.LoginController{}.Logout)
	g.GET("/whoami", controllers.LoginController{}.WhoAmI)
}

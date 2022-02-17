package routers

import (
	"github.com/gin-gonic/gin"
	"web/controllers"
)

func CourseRouterInit(c *gin.Engine) {
	g := c.Group("/api/v1/course")
	g.POST("/create", controllers.CourseController{}.Create)
	g.GET("/get", controllers.CourseController{}.Get)

	g.POST("/schedule", controllers.CourseController{}.Schedule)

}

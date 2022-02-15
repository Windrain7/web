package routers

import (
	"github.com/gin-gonic/gin"
	"web/controllers"
)

func StudentRouterInit(c *gin.Engine) {
	controllers.StudentControllerInit() //缓存预热等操作
	g := c.Group("/api/v1/student")
	g.POST("/book_course", controllers.StudentController{}.BookCourse)
	g.GET("/course", controllers.StudentController{}.Course)
}

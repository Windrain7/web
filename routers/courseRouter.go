package routers

import (
	"github.com/gin-gonic/gin"
	"web/controllers"
)

func CourseRouterInit(c *gin.Engine) {
	g := c.Group("/api/v1/course")
	g.POST("/create", controllers.CourseController{}.Create)
	g.GET("/get", controllers.CourseController{}.Get)

	g.POST("/bind_course", controllers.CourseController{}.BindCourse)
	g.POST("/unbind_course", controllers.CourseController{}.UnbindCourse)
	g.GET("/get_course", controllers.CourseController{}.GetCourse)
	g.POST("/schedule", controllers.CourseController{}.Schedule)

}

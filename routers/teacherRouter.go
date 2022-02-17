package routers

import (
	"github.com/gin-gonic/gin"
	"web/controllers"
)

func TeacherRouterInit(c *gin.Engine) {
	g := c.Group("/api/v1/teacher")
	g.POST("/bind_course", controllers.CourseController{}.BindCourse)
	g.POST("/unbind_course", controllers.CourseController{}.UnbindCourse)
	g.GET("/get_course", controllers.CourseController{}.GetCourse)
}

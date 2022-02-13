package controllers

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"web/models"
)

type StudentController struct{}

const StudentPrefix = "student_id_"
const COURSE_PREFIX = "course_id_"

func StudentControllerInit() {

}

func containsStudentId(studentId string) bool {
	models.Client.Get(StudentPrefix + studentId)
	return true
}

func (con StudentController) BookCourse(c *gin.Context) {
	var request models.BookCourseRequest
	c.BindJSON(&request)
	if request.StudentID == "" {
		log.Println("解析参数错误")
		c.JSON(http.StatusOK, models.BookCourseResponse{Code: models.ParamInvalid})
	}

}

func (con StudentController) Course(c *gin.Context) {

}

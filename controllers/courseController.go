package controllers

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
	"web/models"
)

type CourseController struct{}

func (con CourseController) Create(c *gin.Context) {
	var request models.CreateCourseRequest
	if err := c.BindJSON(&request); err != nil {
		log.Println("创建课程失败，json参数不合法")
		c.JSON(http.StatusOK, models.CreateCourseResponse{
			Code: models.ParamInvalid,
			Data: struct {
				CourseID string
			}{},
		})
		return
	}
	course := models.Course{
		Name:   request.Name,
		Cap:    request.Cap,
		Remain: request.Cap,
	}
	if err := models.Db.Create(&course).Error; err != nil {
		log.Printf("创建课程失败，插入数据库错误")
		log.Println(err)
		c.JSON(http.StatusOK, models.CreateCourseResponse{
			Code: models.UnknownError,
			Data: struct {
				CourseID string
			}{},
		})
	} else {
		c.JSON(http.StatusOK, models.CreateCourseResponse{
			Code: models.OK,
			Data: struct {
				CourseID string
			}{strconv.FormatInt(course.Id, 10)},
		})
	}

}

func (con CourseController) Get(c *gin.Context) {
	var request models.GetCourseRequest
	if err := c.BindQuery(&request); err != nil {
		log.Println("获取课程失败，url参数不合法")
		c.JSON(http.StatusOK, models.GetCourseResponse{
			Code: models.ParamInvalid,
			Data: models.TCourse{},
		})
		return
	}
	id, _ := strconv.ParseInt(request.CourseID, 10, 64)
	var course models.Course
	models.Db.First(&course, id)
	if course.Id == 0 {
		log.Printf("查找课程失败，课程id:%d不存在\n", id)
		c.JSON(http.StatusOK, models.GetCourseResponse{
			Code: models.CourseNotExisted,
			Data: models.TCourse{},
		})
	} else {
		c.JSON(http.StatusOK, models.GetCourseResponse{
			Code: models.OK,
			Data: models.TCourse{
				CourseID:  c.Query("CourseID"),
				Name:      course.Name,
				TeacherID: strconv.FormatInt(course.TeacherId, 10),
			},
		})
	}
}

// TODO 没有检查老师是否存在以及是否对应id是否是老师
func (con CourseController) BindCourse(c *gin.Context) {
	var request models.BindCourseRequest
	if err := c.BindJSON(&request); err != nil {
		log.Println("绑定课程失败，JSON参数不合法")
		c.JSON(http.StatusOK, models.BindCourseResponse{Code: models.ParamInvalid})
		return
	}
	id, _ := strconv.ParseInt(request.CourseID, 10, 64)
	teacherId, _ := strconv.ParseInt(request.TeacherID, 10, 64)
	var course models.Course
	if models.Db.First(&course, id); course.Id == 0 {
		log.Printf("绑定失败，课程id:%d不存在\n", id)
		c.JSON(http.StatusOK, models.BindCourseResponse{Code: models.CourseNotExisted})
	} else if course.TeacherId != 0 {
		log.Printf("绑定失败，课程id:%d的课程已绑定\n", id)
		c.JSON(http.StatusOK, models.BindCourseResponse{Code: models.CourseHasBound})
	} else {
		course.TeacherId = teacherId
		models.Db.Save(&course)
		c.JSON(http.StatusOK, models.BindCourseResponse{Code: models.OK})
	}
}

func (con CourseController) UnbindCourse(c *gin.Context) {
	var request models.UnbindCourseRequest
	if err := c.BindJSON(&request); err != nil {
		log.Println("解绑课程失败，json参数不合法")
		c.JSON(http.StatusOK, models.UnbindCourseResponse{Code: models.ParamInvalid})
		return
	}
	id, _ := strconv.ParseInt(request.CourseID, 10, 64)
	teacherId, _ := strconv.ParseInt(request.TeacherID, 10, 64)
	var course models.Course
	if models.Db.First(&course, id); course.Id == 0 {
		log.Printf("解绑失败，课程id:%d不存在\n", id)
		c.JSON(http.StatusOK, models.UnbindCourseResponse{Code: models.CourseNotExisted})
	} else if course.TeacherId == 0 || course.TeacherId != teacherId {
		log.Printf("解绑失败，课程id:%d的课程没有绑定老师id:%d的老师\n", id, teacherId)
		c.JSON(http.StatusOK, models.UnbindCourseResponse{Code: models.CourseNotBind})
	} else {
		course.TeacherId = 0
		models.Db.Save(&course)
		c.JSON(http.StatusOK, models.UnbindCourseResponse{Code: models.OK})
	}
}

//TODO 这个返回值有问题吧，返回指针前端啥也看不出啊
func (con CourseController) GetCourse(c *gin.Context) {
	var request models.GetTeacherCourseRequest
	if err := c.BindQuery(&request); err != nil {
		log.Println("查看课程失败，url参数不合法")
		c.JSON(http.StatusOK, models.GetTeacherCourseResponse{Code: models.ParamInvalid})
		return
	}
	teacherId, _ := strconv.ParseInt(request.TeacherID, 10, 64)
	var courses []models.Course
	models.Db.Where("teacher_id=?", teacherId).Find(&courses)
	tCourses := make([]*models.TCourse, len(courses))
	for i, course := range courses {
		tCourses[i] = &models.TCourse{
			CourseID:  strconv.FormatInt(course.Id, 10),
			Name:      course.Name,
			TeacherID: strconv.FormatInt(course.TeacherId, 10),
		}
	}
	c.JSON(http.StatusOK, models.GetTeacherCourseResponse{
		Code: models.OK,
		Data: struct {
			CourseList []*models.TCourse
		}{tCourses},
	})
}

func (con CourseController) Schedule(c *gin.Context) {
	var request models.ScheduleCourseRequest
	if err := c.BindJSON(&request); err != nil {
		log.Println("求解课程失败，Json参数不合法")
		c.JSON(http.StatusOK, models.ScheduleCourseResponse{Code: models.ParamInvalid})
		return
	}
	relation := request.TeacherCourseRelationShip
	match := make(map[string]string) //课程对应老师
	vis := make(map[string]bool)     //标记表
	//找增广路径
	var dfs func(string) bool
	dfs = func(i string) bool {
		var list []string
		if list = relation[i]; list == nil {
			return false
		}
		for _, course := range list {
			if vis[course] {
				continue
			}
			vis[course] = true
			if match[course] == "" || dfs(match[course]) {
				match[course] = i
				return true
			}
		}
		return false
	}

	for teacher := range relation {
		dfs(teacher)
		vis = make(map[string]bool)
	}

	//转变为老师对应课程
	result := make(map[string]string)
	for key, val := range match {
		result[val] = key
	}
	c.JSON(http.StatusOK, models.ScheduleCourseResponse{
		Code: models.OK,
		Data: result,
	})
}

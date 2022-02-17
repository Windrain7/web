package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"log"
	"net/http"
	"strconv"
	"web/models"
)

type StudentController struct{}
type bookMessage struct {
	studentId string
	courseId  string
}

var noRemain = map[string]bool{}

const StudentPrefix = "student_"
const CoursePrefix = "course_"
const BookPrefix = "book_"

var bookChannel = make(chan bookMessage, 20000)

func StudentControllerInit() {
	preheat()
	go bookMessageConsumer()
}

func (con StudentController) BookCourse(c *gin.Context) {
	var request models.BookCourseRequest
	if c.BindJSON(&request) != nil {
		log.Println("抢课失败，解析参数错误")
		c.JSON(http.StatusOK, models.BookCourseResponse{Code: models.ParamInvalid})
		return
	}
	if !hasStudent(request.StudentID) {
		log.Println("抢课失败，学生id:", request.StudentID, "不存在")
		c.JSON(http.StatusOK, models.BookCourseResponse{Code: models.StudentNotExisted})
		return
	}
	if !hasCourse(request.CourseID) {
		log.Println("抢课失败，课程id:", request.CourseID, "不存在")
		c.JSON(http.StatusOK, models.BookCourseResponse{Code: models.CourseNotExisted})
		return
	}
	//本地判断是否没有库存
	if noRemain[CoursePrefix+request.CourseID] {
		log.Println("抢课失败，课程id:", request.CourseID, "已满")
		c.JSON(http.StatusOK, models.BookCourseResponse{Code: models.CourseNotAvailable})
		return
	}
	if !tryBookCourse(CoursePrefix+request.CourseID, BookPrefix+request.StudentID, c) {
		return
	}
	//写入mysql
	bookChannel <- bookMessage{courseId: request.CourseID, studentId: request.StudentID}
	c.JSON(http.StatusOK, models.BookCourseResponse{Code: models.OK})
}

func (con StudentController) Course(c *gin.Context) {
	var request models.GetStudentCourseRequest
	if c.BindQuery(&request) != nil {
		log.Println("获取课程失败，解析参数错误")
		c.JSON(http.StatusOK, models.GetStudentCourseResponse{Code: models.ParamInvalid})
		return
	}
	if !hasStudent(request.StudentID) {
		log.Printf("获取课程失败,学生id:%s不存在", request.StudentID)
		c.JSON(http.StatusOK, models.GetStudentCourseResponse{Code: models.StudentNotExisted})
		return
	}
	bookKey := BookPrefix + request.StudentID
	res, _ := models.Rdb.SMembers(models.Ctx, bookKey).Result()
	//没有这个键，去mysql找，并且同步到redis
	if len(res) == 0 {
		var books []models.StudentCourse
		models.Db.Where("student_id=?", request.StudentID).Find(&books)
		//表示没选课
		if len(books) == 0 {
			log.Printf("获取课程失败,学生id:%s没有课程\n", request.StudentID)
			c.JSON(http.StatusOK, models.GetStudentCourseResponse{Code: models.StudentHasNoCourse})
			return
		}
		for _, book := range books {
			//选课表同步到redis
			models.Rdb.SAdd(models.Ctx, bookKey, CoursePrefix+strconv.FormatInt(book.CourseId, 10))
		}
		res, _ = models.Rdb.SMembers(models.Ctx, bookKey).Result()
	}
	//运行到这里res一定有值
	tCourse := make([]models.TCourse, len(res))
	for i, courseKey := range res {
		tCourse[i] = getCourseRedis(courseKey)
	}
	c.JSON(http.StatusOK, models.GetStudentCourseResponse{
		Code: models.OK,
		Data: struct {
			CourseList []models.TCourse
		}{tCourse},
	})
}

//是否有该学生
func hasStudent(studentId string) bool {
	studentKey := StudentPrefix + studentId
	val, err := models.Rdb.Get(models.Ctx, studentKey).Result()
	//redis没有去mysql找，并且设定对应key
	if err == redis.Nil {
		var student models.Member
		id, _ := strconv.ParseInt(studentId, 10, 64)
		models.Db.First(&student, id)
		//存在，没删且是学生
		if student.Id != 0 && student.Deleted != 1 && student.UserType == models.Student {
			models.Rdb.Set(models.Ctx, studentKey, "1", redis.KeepTTL) //1表示存在该学生
			return true
		} else {
			models.Rdb.Set(models.Ctx, studentKey, "0", redis.KeepTTL)
			return false
		}
	} else {
		return val == "1"
	}
}

//是否有该课程
func hasCourse(courseId string) bool {
	courseKey := CoursePrefix + courseId
	val, err := models.Rdb.HGet(models.Ctx, courseKey, "id").Result()
	//redis没有去mysql找，并且设定对应key
	if err == redis.Nil {
		var course models.Course
		id, _ := strconv.ParseInt(courseId, 10, 64)
		models.Db.First(&course, id)
		//存在该课程
		if course.Id != 0 {
			models.Rdb.HSet(models.Ctx, courseKey, map[string]interface{}{
				"id":        course.Id,
				"name":      course.Name,
				"remain":    course.Remain,
				"teacherId": strconv.FormatInt(course.TeacherId, 10),
			})
			return true
		} else {
			models.Rdb.HSet(models.Ctx, courseKey, "id", 0) //表示不存在该课程
			return false
		}
	}
	return val != "0"
}

//从redis中获取课程，如果没有就mysql里找，并且添加到redis里面
func getCourseRedis(courseKey string) models.TCourse {
	id := courseKey[len(CoursePrefix):]
	redisCourse, err := models.Rdb.HGetAll(models.Ctx, courseKey).Result()
	//redis里面没有，去mysql中找
	if err == redis.Nil {
		var course models.Course
		models.Db.First(&course, id)
		models.Rdb.HSet(models.Ctx, courseKey, map[string]interface{}{
			"id":        course.Id,
			"name":      course.Name,
			"remain":    course.Remain,
			"teacherId": strconv.FormatInt(course.TeacherId, 10),
		})
		return models.TCourse{
			CourseID:  id,
			Name:      course.Name,
			TeacherID: strconv.FormatInt(course.TeacherId, 10),
		}
	}
	return models.TCourse{
		CourseID:  id,
		Name:      redisCourse["name"],
		TeacherID: redisCourse["teacherId"],
	}
}

func tryBookCourse(courseIdKey, bookCourseKey string, c *gin.Context) bool {
	res, err := models.Rdb.EvalSha(models.Ctx, models.BookSHA, []string{courseIdKey, bookCourseKey}).Result()
	if err != nil {
		log.Println("抢课失败，执行lua脚本错误")
		c.JSON(http.StatusOK, models.BookCourseResponse{Code: models.UnknownError})
		return false
	}
	code, ok := res.(int64)
	if !ok {
		log.Println("抢课失败，lua脚本返回值错误")
		c.JSON(http.StatusOK, models.BookCourseResponse{Code: models.UnknownError})
		return false
	}
	if code == -1 {
		log.Printf("抢课失败，%s课程容量不足\n", courseIdKey)
		noRemain[courseIdKey] = true
		c.JSON(http.StatusOK, models.BookCourseResponse{Code: models.CourseNotAvailable})
		return false
	} else if code == -2 {
		log.Printf("抢课失败，%s已经选过该课程\n", bookCourseKey)
		c.JSON(http.StatusOK, models.BookCourseResponse{Code: models.StudentHasCourse})
		return false
	}
	return true
}

func bookMessageConsumer() {
	for true {
		message := <-bookChannel
		sId, _ := strconv.ParseInt(message.studentId, 10, 64)
		cId, _ := strconv.ParseInt(message.courseId, 10, 64)
		models.Db.Exec("update course set remain = remain-1 WHERE id=?", cId)
		models.Db.Create(&models.StudentCourse{
			StudentId: sId,
			CourseId:  cId,
		})
	}
}

func preheat() {
	//加入所有学生
	var students []models.Member
	models.Db.Where("user_type=? and deleted=0", models.Student).Find(&students)
	for _, student := range students {
		models.Rdb.Set(models.Ctx, StudentPrefix+strconv.FormatInt(student.Id, 10), "1", redis.KeepTTL)
	}
	//加入所有课程
	var courses []models.Course
	models.Db.Find(&courses)
	for _, course := range courses {
		models.Rdb.HSet(models.Ctx, CoursePrefix+strconv.FormatInt(course.Id, 10), map[string]interface{}{
			"id":        course.Id,
			"name":      course.Name,
			"remain":    course.Remain,
			"teacherId": strconv.FormatInt(course.TeacherId, 10),
		})
	}
	//加入所有选课关系
	var books []models.StudentCourse
	models.Db.Find(&books)
	for _, book := range books {
		models.Rdb.SAdd(models.Ctx, BookPrefix+strconv.FormatInt(book.StudentId, 10), CoursePrefix+strconv.FormatInt(book.CourseId, 10))
	}
}

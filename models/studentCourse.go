package models

type StudentCourse struct {
	Id        int64
	StudentId int64
	CourseId  int64
}

func (s StudentCourse) TableName() string {
	return "student_course"
}

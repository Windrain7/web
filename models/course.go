package models

type Course struct {
	Id        int64
	Name      string
	Cap       int
	Count     int
	TeacherId int64
}

func (c Course) TableName() string {
	return "course"
}

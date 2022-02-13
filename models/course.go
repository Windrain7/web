package models

type Course struct {
	Id        int64
	Name      string
	Cap       int
	Remain    int
	TeacherId int64
}

func (c Course) TableName() string {
	return "course"
}

package models

type Member struct {
	Id       int64
	Username string
	Password string
	Nickname string
	UserType UserType
	Deleted  int8
}

func (m Member) TableName() string {
	return "member"
}

package controllers

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
)
import "web/models"

type MemberController struct{}

func (con MemberController) Index(c *gin.Context) {

}

func (con MemberController) Create(c *gin.Context) {
	//TODO 检查用户是否为管理员
	var member models.Member
	var res models.CreateMemberResponse
	username := c.Query("Username")
	if len(username) < 8 || len(username) > 20 {
		paraInvalidResponse(c)
		return
	}
	models.Db.Where("username = ?", username).First(&member)
	//存在该用户名且没有删除该用户
	if member.Id != 0 && member.Deleted == 0 {
		log.Printf("存在用户%s\n", username)
		res.Code = models.UserHasExisted
		res.Data.UserID = "-1"
		c.JSON(http.StatusOK, res)
		return
	}
	password := c.Query("Password")
	if !validPass(password) {
		paraInvalidResponse(c)
		return
	}
	nickname := c.Query("Nickname")
	if len(nickname) < 4 || len(nickname) > 20 {
		paraInvalidResponse(c)
		return
	}
	userType, err := strconv.Atoi(c.Query("UserType"))
	if err != nil || userType > 3 || userType < 1 {
		paraInvalidResponse(c)
		return
	}
	member.Nickname = nickname
	member.Password = password
	member.Username = username
	member.UserType = int8(userType)
	result := models.Db.Create(&member)
	if result.Error != nil {
		log.Println(result.Error)
		res.Code = models.UnknownError
		res.Data.UserID = "-1"
		c.JSON(http.StatusOK, res)
		return
	}
	res.Code = models.OK
	res.Data.UserID = strconv.Itoa(int(member.Id))
	c.JSON(http.StatusOK, res)
}

func paraInvalidResponse(c *gin.Context) {
	log.Println("参数错误")
	res := models.CreateMemberResponse{
		Code: models.ParamInvalid,
		Data: struct {
			UserID string
		}{"-1"},
	}
	c.JSON(http.StatusOK, res)
}

func validPass(password string) bool {
	if len(password) < 8 || len(password) > 20 {
		return false
	}
	var hasCapital, hasLower, hasDigital bool
	for _, ch := range password {
		hasLower = (ch >= 'a' && ch <= 'z') || hasLower
		hasCapital = (ch >= 'A' && ch <= 'Z') || hasCapital
		hasDigital = (ch >= '0' && ch <= '9') || hasDigital
	}
	return hasDigital && hasLower && hasCapital
}

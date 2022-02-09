package controllers

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
	"web/models"
)

type LoginController struct{}

func (con LoginController) Login(c *gin.Context) {
	var request models.LoginRequest
	if err := c.BindJSON(&request); err != nil {
		log.Println("登陆失败，密码错误")
		log.Println("登陆失败，绑定json错误")
		c.JSON(http.StatusOK, models.LoginResponse{
			Code: models.WrongPassword,
			Data: struct {
				UserID string
			}{"-1"},
		})
		return
	}
	//检查输入用户名和密码的合法性
	if len(request.Username) < 8 || len(request.Username) > 20 || !ValidPass(request.Password) {
		log.Printf("登陆失败，用户名:%s, 密码:%s\n", request.Username, request.Password)
		c.JSON(http.StatusOK, models.LoginResponse{
			Code: models.WrongPassword,
			Data: struct {
				UserID string
			}{"-1"},
		})
	}
	var member models.Member
	models.Db.Where("username=?", request.Username).First(&member)
	//用户不存在，密码错误，用户已删除全部按密码错误处理
	if member.Id == 0 || member.Password != request.Password || member.Deleted == 1 {
		log.Println("登陆失败，用户不存在或者密码错误或者用户已删除")
		c.JSON(http.StatusOK, models.LoginResponse{
			Code: models.WrongPassword,
			Data: struct {
				UserID string
			}{"-1"},
		})
	} else {
		//cookie为 member.UserType + member.Id
		//UserType 1为管理员，2为学生，3为教师
		t := strconv.FormatInt(int64(member.UserType), 10) + strconv.FormatInt(member.Id, 10)
		c.SetCookie("camp-session", t, 3600, "/", "180.184.74.66", false, true)
		log.Printf("当前用户cookie为:%s\n", t)
		c.JSON(http.StatusOK, models.LoginResponse{
			Code: models.OK,
			Data: struct {
				UserID string
			}{strconv.FormatInt(member.Id, 10)},
		})
	}
}

func (con LoginController) Logout(c *gin.Context) {
	val, err := c.Cookie("camp-session")
	//未登录
	if err != nil {
		log.Println("登出失败，用户未登录")
		c.JSON(http.StatusOK, models.LogoutResponse{Code: models.LoginRequired})
		return
	}
	c.SetCookie("camp-session", val, -1, "/", "180.184.74.66", false, true)
	c.JSON(http.StatusOK, models.LogoutResponse{Code: models.OK})
}

// WhoAmI TODO 会出现这里登录着，但是被别人删了的情况吗
func (con LoginController) WhoAmI(c *gin.Context) {
	//未登录
	val, err := c.Cookie("camp-session")
	if err != nil {
		log.Println("查看信息失败，用户未登录")
		c.JSON(http.StatusOK, models.WhoAmIResponse{
			Code: models.LoginRequired,
			Data: models.TMember{},
		})
		return
	}
	id, _ := strconv.ParseInt(val[1:], 10, 64)
	var member models.Member
	models.Db.Find(&member, id)
	c.JSON(http.StatusOK, models.WhoAmIResponse{
		Code: models.OK,
		Data: models.TMember{
			UserID:   strconv.FormatInt(member.Id, 10),
			Nickname: member.Nickname,
			Username: member.Username,
			UserType: models.UserType(member.UserType),
		},
	})

}

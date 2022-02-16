package controllers

import (
	"github.com/gin-contrib/sessions"
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
		return
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
		session := sessions.Default(c)
		session.Set("member", member)
		session.Save()
		c.JSON(http.StatusOK, models.LoginResponse{
			Code: models.OK,
			Data: struct {
				UserID string
			}{strconv.FormatInt(member.Id, 10)},
		})
	}
}

func (con LoginController) Logout(c *gin.Context) {
	session := sessions.Default(c)
	if session.Get("member") == nil {
		log.Println("登出失败，用户未登录")
		c.JSON(http.StatusOK, models.LogoutResponse{Code: models.LoginRequired})
		return
	}
	session.Delete("member")
	session.Save()
	c.JSON(http.StatusOK, models.LogoutResponse{Code: models.OK})
}

func (con LoginController) WhoAmI(c *gin.Context) {
	session := sessions.Default(c)
	if session.Get("member") == nil {
		log.Println("查看用户信息失败，用户未登录")
		c.JSON(http.StatusOK, models.LogoutResponse{Code: models.LoginRequired})
		return
	}
	member := session.Get("member").(models.Member)
	c.JSON(http.StatusOK, models.WhoAmIResponse{
		Code: models.OK,
		Data: models.TMember{
			UserID:   strconv.FormatInt(member.Id, 10),
			Nickname: member.Nickname,
			Username: member.Username,
			UserType: member.UserType,
		},
	})
}

package controllers

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
	"web/models"
)

type MemberController struct{}

func (con MemberController) Index(c *gin.Context) {
	var member models.Member
	id, _ := strconv.Atoi(c.Query("UserID"))
	models.Db.First(&member, id)
	if member.Id == 0 {
		log.Printf("查找用户ID:%d失败，不存在该用户\n", id)
		c.JSON(http.StatusOK, models.GetMemberResponse{
			Code: models.UserNotExisted,
			Data: models.TMember{},
		})
	} else if member.Deleted == 1 {
		log.Printf("查找用户ID:%d失败，用户已删除\n")
		c.JSON(http.StatusOK, models.GetCourseResponse{
			Code: models.UserHasDeleted,
			Data: models.TCourse{},
		})
	} else {
		res := models.GetMemberResponse{
			Code: models.OK,
			Data: models.TMember{
				UserID:   strconv.Itoa(int(member.Id)),
				Nickname: member.Nickname,
				Username: member.Username,
				UserType: models.UserType(member.UserType),
			},
		}
		c.JSON(http.StatusOK, res)
	}
}

// TODO 需要细分为未登录和没有权限吗
func (con MemberController) Create(c *gin.Context) {
	//检查用户是否为管理员
	val, err := c.Cookie("camp-session")
	//没登陆或者不是管理员
	if err != nil || val[:1] != "1" {
		c.JSON(http.StatusOK, models.CreateMemberResponse{
			Code: models.PermDenied,
			Data: struct {
				UserID string
			}{},
		})
		return
	}
	var member models.Member
	var res models.CreateMemberResponse
	username := c.Query("Username")
	if len(username) < 8 || len(username) > 20 {
		paraInvalidResponse(c, "用户名")
		return
	}
	models.Db.Where("username = ?", username).First(&member)
	//存在该用户名且没有删除该用户
	if member.Id != 0 && member.Deleted == 0 {
		log.Printf("创建用户失败，已存在用户%s\n", username)
		res.Code = models.UserHasExisted
		res.Data.UserID = "-1"
		c.JSON(http.StatusOK, res)
		return
	}
	password := c.Query("Password")
	if !validPass(password) {
		paraInvalidResponse(c, "密码")
		return
	}
	nickname := c.Query("Nickname")
	if len(nickname) < 4 || len(nickname) > 20 {
		paraInvalidResponse(c, "昵称")
		return
	}
	userType, err := strconv.Atoi(c.Query("UserType"))
	if err != nil || userType > 3 || userType < 1 {
		paraInvalidResponse(c, "用户类型")
		return
	}
	member.Nickname = nickname
	member.Password = password
	member.Username = username
	member.UserType = int8(userType)
	result := models.Db.Create(&member)
	if result.Error != nil {
		log.Print("数据库创建用户失败")
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

func paraInvalidResponse(c *gin.Context, t string) {
	log.Printf("创建用户失败，参数%s错误\n", t)
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

func (con MemberController) List(c *gin.Context) {
	var members []models.Member
	offset, _ := strconv.Atoi(c.Query("Offset"))
	limit, _ := strconv.Atoi(c.Query("Limit"))
	models.Db.Limit(limit).Offset(offset).Find(&members)
	tMembers := make([]models.TMember, len(members))
	for i, member := range members {
		tMembers[i] = models.TMember{
			UserID:   strconv.Itoa(int(member.Id)),
			Nickname: member.Nickname,
			Username: member.Username,
			UserType: models.UserType(member.UserType),
		}
	}
	c.JSON(http.StatusOK, models.GetMemberListResponse{
		Code: models.OK,
		Data: struct {
			MemberList []models.TMember
		}{tMembers},
	})
}

func (con MemberController) Update(c *gin.Context) {
	var member models.Member
	id, err := strconv.Atoi(c.Query("UserID"))
	nickname := c.Query("Nickname")
	if err != nil || len(nickname) < 4 || len(nickname) > 8 {
		log.Printf("更新昵称失败，更新后的昵称:%s不合法\n", nickname)
		c.JSON(http.StatusOK, models.ParamInvalid)
		return
	}
	models.Db.First(&member, id)
	if member.Id == 0 {
		log.Printf("更新昵称失败，用户id:%d不存在\n", id)
		c.JSON(http.StatusOK, models.UserNotExisted)
	} else if member.Deleted == 1 {
		log.Printf("更新昵称失败，用户id:%d已删除\n", id)
		c.JSON(http.StatusOK, models.UserHasDeleted)
	} else {
		member.Nickname = nickname
		models.Db.Save(&member)
		c.JSON(http.StatusOK, models.OK)
	}
}

func (con MemberController) Delete(c *gin.Context) {
	var member models.Member
	id, _ := strconv.Atoi(c.Query("UserID"))
	models.Db.First(&member, id)
	if member.Id == 0 {
		log.Printf("删除用户失败,用户id:%d不存在\n", id)
		c.JSON(http.StatusOK, models.UserNotExisted)
	} else if member.Deleted == 1 {
		log.Printf("删除用户失败,用户id:%d已删除\n", id)
		c.JSON(http.StatusOK, models.UserHasDeleted)
	} else {
		member.Deleted = 1
		models.Db.Save(&member)
		c.JSON(http.StatusOK, models.OK)
	}
}

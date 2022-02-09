package controllers

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
	"web/models"
)

type MemberController struct{}

// 没有区分用户是否登录
func (con MemberController) Index(c *gin.Context) {
	var member models.Member
	id, _ := strconv.ParseInt(c.Query("UserID"), 10, 64)
	models.Db.First(&member, id)
	if member.Id == 0 {
		log.Printf("查找用户ID:%d失败，不存在该用户\n", id)
		c.JSON(http.StatusOK, models.GetMemberResponse{
			Code: models.UserNotExisted,
			Data: models.TMember{},
		})
	} else if member.Deleted == 1 {
		log.Printf("查找用户ID:%d失败，用户已删除\n", id)
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
	if !canCreate(c) {
		return
	}
	var member models.Member
	var res models.CreateMemberResponse
	username := c.Query("Username")
	if len(username) < 8 || len(username) > 20 {
		paraInvalidResponse(c, "用户名")
		return
	}
	models.Db.Where("username=?", username).First(&member)
	//存在该用户名且没有删除该用户
	if member.Id != 0 && member.Deleted == 0 {
		log.Printf("创建用户失败，已存在用户%s\n", username)
		res.Code = models.UserHasExisted
		res.Data.UserID = "-1"
		c.JSON(http.StatusOK, res)
		return
	}
	password := c.Query("Password")
	if !ValidPass(password) {
		paraInvalidResponse(c, "密码")
		return
	}
	nickname := c.Query("Nickname")
	if len(nickname) < 4 || len(nickname) > 20 {
		paraInvalidResponse(c, "昵称")
		return
	}
	userType, err := strconv.ParseInt(c.Query("UserType"), 10, 8)
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

//用户可以创建成员返回true，否则返回对应json
func canCreate(c *gin.Context) bool {
	t := GetUserTypeFromCookie(c)
	// 未登录
	if t == 0 {
		println("创建用户失败,用户未登录")
		c.JSON(http.StatusOK, models.CreateMemberResponse{
			Code: models.LoginRequired,
			Data: struct {
				UserID string
			}{},
		})
		return false
	}
	// 无权限
	if t != 1 {
		println("创建用户失败,用户无权限")
		c.JSON(http.StatusOK, models.CreateMemberResponse{
			Code: models.PermDenied,
			Data: struct {
				UserID string
			}{},
		})
		return false
	}
	return true
}

// 返回创建成员参数错误
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

// TODO 需要判断是否登录和和没有权限吗
func (con MemberController) List(c *gin.Context) {
	var members []models.Member
	offset, _ := strconv.Atoi(c.Query("Offset"))
	limit, _ := strconv.Atoi(c.Query("Limit"))
	//过滤已删除的成员
	models.Db.Where("deleted=0").Limit(limit).Offset(offset).Find(&members)
	tMembers := make([]models.TMember, len(members))
	//使用Member构造TMember
	for i, member := range members {
		tMembers[i] = models.TMember{
			UserID:   strconv.FormatInt(member.Id, 10),
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

// TODO 要求是 非管理员只能改自己的nickname，管理员可以改所有人的nickname吗
// TODO 要是用户额外传入了密码参数怎么处理
func (con MemberController) Update(c *gin.Context) {
	userType := GetUserTypeFromCookie(c)
	// 未登录
	if userType == 0 {
		log.Println("更新用户失败，用户未登录")
		c.JSON(http.StatusOK, models.UpdateMemberResponse{Code: models.LoginRequired})
	}
	userId := GetIdFromCookie(c)
	id, _ := strconv.ParseInt(c.Query("UserID"), 10, 64)
	// 不是管理员并且改的不是自己id
	if userId != id && userType != 1 {
		log.Println("更新用户失败，用户无权限")
		c.JSON(http.StatusOK, models.UpdateMemberResponse{Code: models.PermDenied})
	}
	nickname := c.Query("Nickname")
	if len(nickname) < 4 || len(nickname) > 8 {
		log.Printf("更新昵称失败，更新后的昵称:%s不合法\n", nickname)
		c.JSON(http.StatusOK, models.UpdateMemberResponse{Code: models.ParamInvalid})
		return
	}
	var member models.Member
	models.Db.First(&member, id)
	if member.Id == 0 {
		log.Printf("更新昵称失败，用户id:%d不存在\n", id)
		c.JSON(http.StatusOK, models.UpdateMemberResponse{Code: models.UserNotExisted})
	} else if member.Deleted == 1 {
		log.Printf("更新昵称失败，用户id:%d已删除\n", id)
		c.JSON(http.StatusOK, models.UpdateMemberResponse{Code: models.UserHasDeleted})
	} else {
		member.Nickname = nickname
		models.Db.Save(&member)
		c.JSON(http.StatusOK, models.UpdateMemberResponse{Code: models.OK})
	}
}

// TODO 需要判断是否登录和没有权限吗
func (con MemberController) Delete(c *gin.Context) {
	var member models.Member
	id, _ := strconv.ParseInt(c.Query("UserID"), 10, 64)
	models.Db.First(&member, id)
	if member.Id == 0 {
		log.Printf("删除用户失败,用户id:%d不存在\n", id)
		c.JSON(http.StatusOK, models.DeleteMemberResponse{Code: models.UserNotExisted})
	} else if member.Deleted == 1 {
		log.Printf("删除用户失败,用户id:%d已删除\n", id)
		c.JSON(http.StatusOK, models.DeleteMemberResponse{Code: models.UserHasDeleted})
	} else {
		member.Deleted = 1
		models.Db.Save(&member)
		c.JSON(http.StatusOK, models.DeleteMemberResponse{Code: models.OK})
	}
}

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
	var request models.GetMemberRequest
	if err := c.BindQuery(&request); err != nil {
		log.Println("查找用户失败，参数错误")
		c.JSON(http.StatusOK, models.GetMemberResponse{
			Code: models.ParamInvalid,
			Data: models.TMember{},
		})
		return
	}
	id, _ := strconv.ParseInt(request.UserID, 10, 64)
	var member models.Member
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
				UserType: member.UserType,
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
	var request models.CreateMemberRequest
	var res models.CreateMemberResponse
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Println("创建用户失败，参数解析错误")
		res.Code = models.ParamInvalid
		c.JSON(http.StatusOK, res)
		return
	}
	if len(request.Username) < 8 || len(request.Username) > 20 {
		paraInvalidResponse(c, "用户名"+request.Username)
		return
	}
	var member models.Member
	models.Db.Where("username=?", request.Username).First(&member)
	//存在该用户名
	if member.Id != 0 {
		log.Printf("创建用户失败，已存在用户%s\n", request.Username)
		res.Code = models.UserHasExisted
		res.Data.UserID = "-1"
		c.JSON(http.StatusOK, res)
		return
	}
	if !ValidPass(request.Password) {
		paraInvalidResponse(c, "密码"+request.Password)
		return
	}
	if len(request.Nickname) < 4 || len(request.Nickname) > 20 {
		paraInvalidResponse(c, "昵称"+request.Nickname)
		return
	}
	if request.UserType > 3 || request.UserType < 1 {
		paraInvalidResponse(c, "用户类型"+strconv.Itoa(int(request.UserType)))
		return
	}
	member.Username = request.Username
	member.Password = request.Password
	member.Nickname = request.Nickname
	member.UserType = request.UserType
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
	val, err := c.Cookie("camp-session")
	if err != nil || val[:1] != "1" {
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

func (con MemberController) List(c *gin.Context) {
	var request models.GetMemberListRequest
	if err := c.BindQuery(&request); err != nil {
		log.Println("列表查询失败，参数解析错误")
		c.JSON(http.StatusOK, models.GetMemberListResponse{
			Code: models.ParamInvalid,
			Data: struct {
				MemberList []models.TMember
			}{},
		})
		return
	}
	var members []models.Member
	log.Println("offset:", request.Offset, "\tlimit:", request.Limit)
	//过滤已删除的成员
	models.Db.Where("deleted=0").Limit(request.Limit).Offset(request.Offset).Find(&members)
	tMembers := make([]models.TMember, len(members))
	//使用Member构造TMember
	for i, member := range members {
		tMembers[i] = models.TMember{
			UserID:   strconv.FormatInt(member.Id, 10),
			Nickname: member.Nickname,
			Username: member.Username,
			UserType: member.UserType,
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
	var request models.UpdateMemberRequest
	if err := c.BindJSON(&request); err != nil {
		log.Println("更新昵称失败，传入参数不合法")
		c.JSON(http.StatusOK, models.UpdateMemberResponse{Code: models.ParamInvalid})
		return
	}
	if len(request.Nickname) < 4 || len(request.Nickname) > 8 {
		log.Printf("更新昵称失败，更新后的昵称:%s不合法\n", request.Nickname)
		c.JSON(http.StatusOK, models.UpdateMemberResponse{Code: models.ParamInvalid})
		return
	}
	var member models.Member
	id, _ := strconv.ParseInt(request.UserID, 10, 64)
	models.Db.First(&member, id)
	if member.Id == 0 {
		log.Printf("更新昵称失败，用户id:%d不存在\n", id)
		c.JSON(http.StatusOK, models.UpdateMemberResponse{Code: models.UserNotExisted})
	} else if member.Deleted == 1 {
		log.Printf("更新昵称失败，用户id:%d已删除\n", id)
		c.JSON(http.StatusOK, models.UpdateMemberResponse{Code: models.UserHasDeleted})
	} else {
		member.Nickname = request.Nickname
		models.Db.Save(&member)
		c.JSON(http.StatusOK, models.UpdateMemberResponse{Code: models.OK})
	}
}

func (con MemberController) Delete(c *gin.Context) {
	var request models.DeleteMemberRequest
	if err := c.BindJSON(&request); err != nil {
		log.Println("删除失败，传入参数不合法")
		c.JSON(http.StatusOK, models.DeleteMemberResponse{Code: models.ParamInvalid})
		return
	}
	id, _ := strconv.ParseInt(request.UserID, 10, 64)
	var member models.Member
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

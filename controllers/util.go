package controllers

import (
	"github.com/gin-gonic/gin"
	"strconv"
	"web/models"
)

func ValidPass(password string) bool {
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

func ValidUsername(username string) bool {
	if len(username) < 8 || len(username) > 20 {
		return false
	}
	for _, ch := range username {
		if (ch < 'a' || ch > 'z') && (ch < 'A' || ch > 'Z') {
			return false
		}
	}
	return true
}

//从cookie获取用户身份，若未登录返回0，登录返回身份
func GetUserTypeFromCookie(c *gin.Context) models.UserType {
	val, err := c.Cookie("camp-session")
	if err != nil {
		return 0
	} else {
		t, _ := strconv.Atoi(val[:1])
		return models.UserType(t)
	}
}

//从cookie获取用户id，若未登录返回0，登录返回id
func GetIdFromCookie(c *gin.Context) int64 {
	val, err := c.Cookie("camp-session")
	if err != nil {
		return 0
	} else {
		t, _ := strconv.ParseInt(val[1:], 10, 64)
		return t
	}
}

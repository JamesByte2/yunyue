package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type loginIn struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func login(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var in loginIn
		if err := c.ShouldBindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var u User
		if err := db.Where("email = ?", in.Email).First(&u).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "邮箱或密码错误"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器错误"})
			return
		}
		if !checkPassword(u.PasswordHash, in.Password) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "邮箱或密码错误"})
			return
		}
		token, err := makeToken(secret, u.ID, u.Role)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "签发凭证失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"token": token, "role": u.Role, "name": u.Name})
	}
}

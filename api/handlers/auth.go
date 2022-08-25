package handlers

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func LogIn(ctx *gin.Context) {
	var credentials struct {
		Username string `binding:"required"`
		Password string `binding:"required"`
	}

	if err := ctx.Bind(&credentials); err != nil {
		return
	}

	password := os.Getenv("PASSWORD")
	username := os.Getenv("USERNAME")

	if username != credentials.Username {
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"cause": "USER_NOT_FOUND"})
		return
	}
	if password != credentials.Password {
		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"cause": "PASSWORD_INCORRECT"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"token": os.Getenv("TOKEN")})
}

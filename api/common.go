package api

import "github.com/gin-gonic/gin"

func Fail(ctx *gin.Context, statusCode int, message string) {
	ctx.AbortWithStatusJSON(statusCode, gin.H{"error": message})
}

func Success(ctx *gin.Context, statusCode int) {
	ctx.JSON(statusCode, nil)
}

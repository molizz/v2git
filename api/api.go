package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

var password string

func Start(addr string, pwd string) {
	password = pwd

	engine := gin.Default()
	engine.Use(validate)

	projects := engine.Group("/projects")
	{
		projects.POST("/", createProject)
		projects.DELETE("/:namespace/:name/", deleteProject)
	}

	engine.Run(addr) // 最好还是用https
}

func validate(ctx *gin.Context) {
	pwd := ctx.Request.Header.Get("PWD")
	if pwd != password {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
	}
}

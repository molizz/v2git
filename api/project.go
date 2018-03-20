package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/molisoft/v2git/repo"
)

func createProject(ctx *gin.Context) {
	body, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"result": 0, "message": "Body error " + err.Error()})
		return
	}
	ns := &struct {
		Namespace string
	}{}
	err = json.Unmarshal(body, ns)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"result": 0, "message": "The body is not valid " + err.Error()})
		return
	}
	repository, err := repo.New(ns.Namespace)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"result": 0, "message": "repo.New error " + err.Error()})
		return
	}
	err = repository.CreateProject()
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"result": 0, "message": "Create project error " + err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"result": 1})
	return
}

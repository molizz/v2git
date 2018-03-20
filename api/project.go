package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"path"

	"github.com/gin-gonic/gin"
	"github.com/molisoft/v2git/repo"
)

type namespace struct {
	Namespace string
}

func createProject(ctx *gin.Context) {
	body, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		Fail(ctx, http.StatusNotAcceptable, "Body error "+err.Error())
		return
	}
	defer ctx.Request.Body.Close()
	ns := &namespace{}
	err = json.Unmarshal(body, ns)
	if err != nil {
		Fail(ctx, http.StatusNotAcceptable, "The body is not valid "+err.Error())
		return
	}
	repository, err := repo.New(ns.Namespace)
	if err != nil {
		Fail(ctx, http.StatusUnprocessableEntity, "repo.New error "+err.Error())
		return
	}
	err = repository.CreateProject()
	if err != nil {
		Fail(ctx, http.StatusUnprocessableEntity, "Create project error "+err.Error())
		return
	}
	Success(ctx, http.StatusCreated)
	return
}

func deleteProject(ctx *gin.Context) {
	ns := &namespace{
		Namespace: path.Join(ctx.Param("namespace"), ctx.Param("name")),
	}
	repository, err := repo.New(ns.Namespace)
	if err != nil {
		Fail(ctx, http.StatusUnprocessableEntity, "repo.New error "+err.Error())
		return
	}
	err = repository.DeleteProject()
	if err != nil {
		Fail(ctx, http.StatusUnprocessableEntity, "Can't delete "+err.Error())
		return
	}
	Success(ctx, http.StatusNoContent)
	return
}

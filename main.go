/*

 */

package main

import (
	"fmt"
	"os"

	"github.com/molisoft/v2git/auth"
	"github.com/molisoft/v2git/cfg"
	"github.com/molisoft/v2git/git"
)

var config *cfg.Config

func init() {
	var err error
	config, err = cfg.New("config.yml")
	if err != nil {
		fmt.Println("Failed to load config.yml. ", err.Error())
		os.Exit(1)
	}
}

func verify(request_path, username, password string) bool {
	authRequest := &cfg.AuthRequest{
		AuthURL:  config.AuthUrl,
		Path:     request_path,
		Username: username,
		Password: password,
	}
	b, err := auth.Verify(authRequest)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return b
}

func main() {
	addr := fmt.Sprintf(":%d", config.Port)
	g := git.New(addr, config.RepoDir)
	g.RegisterVerify(verify)
	g.Start()

	//select {}
}

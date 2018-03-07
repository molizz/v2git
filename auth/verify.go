package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/molisoft/v2git/cfg"
	"github.com/valyala/fasthttp"
)

func Verify(auth *cfg.AuthRequest) (bool, error) {
	if os.Getenv("ENV") == "dev" {
		return true, nil
	}
	body, err := json.Marshal(auth)
	if err != nil {
		return false, errors.New("Verify: json marshal error " + err.Error())
	}
	var args fasthttp.Args
	args.Set("Content-Type", "application/json")
	code, responseBody, err := fasthttp.Post(body, auth.AuthURL, &args)
	if err != nil {
		return false, errors.New("Verify: request api error " + err.Error())
	}
	if code != 200 {
		return false, errors.New(fmt.Sprintf("Verify: request code :%d", code))
	}

	if err != nil {
		return false, err
	}
	authResp := new(cfg.AuthResponse)
	err = json.Unmarshal(responseBody, authResp)
	if err != nil {
		return false, err
	}
	return authResp.IsOK(), nil
}

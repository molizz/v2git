package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/molisoft/v2git/cfg"
)

func Verify(auth *cfg.AuthRequest) (bool, error) {
	if isDev() {
		return true, nil
	}
	body, err := json.Marshal(auth)
	if err != nil {
		return false, errors.New("Verify: json marshal error " + err.Error())
	}

	resp, err := http.Post(auth.AuthURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return false, errors.New("Verify: request api error " + err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		return false, errors.New(fmt.Sprintf("Verify: request code :%d", resp.StatusCode))
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, errors.New("Verify: Response error " + err.Error())
	}
	authResp := new(cfg.AuthResponse)
	err = json.Unmarshal(respBody, authResp)
	if err != nil {
		return false, err
	}
	return authResp.IsOK(), nil
}

func isDev() bool {
	return os.Getenv("ENV") == "dev"
}

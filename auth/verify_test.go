package auth

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/molisoft/v2git/cfg"
)

func TestVerify(t *testing.T) {
	startFakeHttp(":9090")
	authRequest := &cfg.AuthRequest{
		AuthURL:  "http://127.0.0.1:9090/api/v1/auth",
		Path:     "",
		Username: "moli",
		Password: "123123",
	}
	b, err := Verify(authRequest)
	if err != nil {
		t.Error(err)
	}
	if b {
		t.Log("ok")
	}
}

func httpHandler(w http.ResponseWriter, req *http.Request) {
	authResp := &cfg.AuthResponse{
		Result: 0,
	}
	respBody, _ := json.Marshal(authResp)
	w.Write(respBody)
}

func startFakeHttp(addr string) {
	http.HandleFunc("/api/v1/auth", httpHandler)
	go http.ListenAndServe(addr, nil)
}

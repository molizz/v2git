/*

Base https://github.com/asim/git-http-backend

*/

/*
g := git.New(":8080", "/project/base/dir")
g.RegisterVerify(func(user, pwd string) bool {
	// 验证账号密码的正确
})

g.Start()
*/

package git

import (
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/molisoft/v2git/utils"
	"github.com/valyala/fasthttp"
)

type Service struct {
	Method  string
	Handler func(HandlerReq)
	Rpc     string
}

type Config struct {
	ListenAddr  string
	ProjectBase string
	GitBinPath  string
	UploadPack  bool
	ReceivePack bool
}

type HandlerReq struct {
	r    *fasthttp.RequestCtx
	Rpc  string
	Dir  string
	File string
}

type GitHttp struct {
	config   *Config
	services map[string]Service
	verify   Verification
}

type Verification func(string, string, string) bool

var (
	Realm = "V2Git"
)

// Request handling function

func (this *GitHttp) requestHandler(r *fasthttp.RequestCtx) {
	proto := ""
	if r.Request.Header.IsHTTP11() {
		proto = "HTTP 1.1"
	}
	log.Printf("%s %s %s %s", r.RemoteAddr(), r.Method(), r.URI().Path(), proto)
	if !this.verification(r) {
		this.renderUnauthorized(r)
		return
	}

	for match, service := range this.services {
		re, err := regexp.Compile(match)
		if err != nil {
			log.Print(err)
		}

		if m := re.FindStringSubmatch(utils.BytesToString(r.URI().Path())); m != nil {
			if service.Method != utils.BytesToString(r.Method()) {
				this.renderMethodNotAllowed(r)
				return
			}

			rpc := service.Rpc
			file := strings.Replace(utils.BytesToString(r.Path()), m[1]+"/", "", 1)
			dir, err := this.getGitDir(m[1])
			if err != nil {
				log.Print(err)
				this.renderNotFound(r)
				return
			}

			hr := HandlerReq{r, rpc, dir, file}
			service.Handler(hr)
			return
		}
	}
	this.renderNotFound(r)
	return

}

// 验证账号密码

func (this *GitHttp) verification(r *fasthttp.RequestCtx) bool {
	if this.verify == nil {
		return true
	}
	authField := utils.BytesToString(r.Request.Header.Peek("Authorization"))
	if len(authField) > 0 {
		parts := strings.Split(authField, " ")
		if len(parts) == 2 {
			if parts[0] == "Basic" {
				authBytes, err := base64.StdEncoding.DecodeString(parts[1])
				authString := utils.BytesToString(authBytes)
				if err == nil {
					auths := strings.SplitN(authString, ":", 2)
					username := auths[0]
					password := auths[1]
					request_path := utils.BytesToString(r.Path())
					return this.verify(request_path, username, password)
				}
			}
		}
	}
	r.Response.Header.Set("WWW-Authenticate", fmt.Sprintf("Basic realm=\"%s\"", Realm))
	log.Println("Unauthorized!")
	return false
}

// Actual command handling functions

func (this *GitHttp) serviceRpc(hr HandlerReq) {
	r, rpc, dir := hr.r, hr.Rpc, hr.Dir
	access := this.hasAccess(r, dir, rpc, true)

	if access == false {
		this.renderNoAccess(r)
		return
	}

	r.SetContentType(fmt.Sprintf("application/x-git-%s-result", rpc))
	r.SetStatusCode(http.StatusOK)

	args := []string{rpc, "--stateless-rpc", dir}
	cmd := exec.Command(this.config.GitBinPath, args...)
	cmd.Dir = dir
	in, err := cmd.StdinPipe()
	if err != nil {
		log.Print(err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Print(err)
	}

	err = cmd.Start()
	if err != nil {
		log.Print(err)
	}

	switch utils.BytesToString(r.Request.Header.Peek("Content-Encoding")) {
	case "gzip":
		body_bytes, _ := r.Request.BodyGunzip()
		in.Write(body_bytes)
	default:
		r.Request.BodyWriteTo(in)
	}
	in.Close()
	io.Copy(r, stdout)
	cmd.Wait()
}

func (this *GitHttp) getInfoRefs(hr HandlerReq) {
	r, dir := hr.r, hr.Dir
	service_name := this.getServiceType(r)
	access := this.hasAccess(r, dir, service_name, false)

	if access {
		args := []string{service_name, "--stateless-rpc", "--advertise-refs", "."}
		fmt.Println(args)
		refs := this.gitCommand(dir, args...)

		this.hdrNocache(r)
		r.Response.Header.Set("Content-Type", fmt.Sprintf("application/x-git-%s-advertisement", service_name))
		r.SetStatusCode(http.StatusOK)
		r.Write(this.packetWrite("# service=git-" + service_name + "\n"))
		r.Write(this.packetFlush())
		r.Write(refs)
	} else {
		this.updateServerInfo(dir)
		this.hdrNocache(r)
		this.sendFile("text/plain; charset=utf-8", hr)
	}
}

func (this *GitHttp) getInfoPacks(hr HandlerReq) {
	this.hdrCacheForever(hr.r)
	this.sendFile("text/plain; charset=utf-8", hr)
}

func (this *GitHttp) getLooseObject(hr HandlerReq) {
	this.hdrCacheForever(hr.r)
	this.sendFile("application/x-git-loose-object", hr)
}

func (this *GitHttp) getPackFile(hr HandlerReq) {
	this.hdrCacheForever(hr.r)
	this.sendFile("application/x-git-packed-objects", hr)
}

func (this *GitHttp) getIdxFile(hr HandlerReq) {
	this.hdrCacheForever(hr.r)
	this.sendFile("application/x-git-packed-objects-toc", hr)
}

func (this *GitHttp) getTextFile(hr HandlerReq) {
	this.hdrNocache(hr.r)
	this.sendFile("text/plain", hr)
}

// Logic helping functions

func (this *GitHttp) sendFile(content_type string, hr HandlerReq) {
	r := hr.r
	req_file := path.Join(hr.Dir, hr.File)

	f, err := os.Stat(req_file)
	if os.IsNotExist(err) {
		this.renderNotFound(r)
		return
	}
	r.Response.Header.Set("Content-Type", content_type)
	r.Response.Header.Set("Content-Length", fmt.Sprintf("%d", f.Size()))
	r.Response.Header.Set("Last-Modified", f.ModTime().Format(http.TimeFormat))
	fasthttp.ServeFile(r, req_file)
}

func (this *GitHttp) getGitDir(uriPath string) (string, error) {
	root := this.config.ProjectBase

	if root == "" {
		cwd, err := os.Getwd()
		if err != nil {
			log.Print(err)
			return "", err
		}
		root = cwd
	}
	dir, err := UrlToDirPath(uriPath)
	if err != nil {
		return "", err
	}

	f := path.Join(root, dir)
	if _, err := os.Stat(f); os.IsNotExist(err) {
		return "", err
	}
	return f, nil
}

func (this *GitHttp) getServiceType(r *fasthttp.RequestCtx) string {
	service_type := utils.BytesToString(r.FormValue("service"))

	if s := strings.HasPrefix(service_type, "git-"); !s {
		return ""
	}

	return strings.Replace(service_type, "git-", "", 1)
}

func (this *GitHttp) hasAccess(r *fasthttp.RequestCtx, dir string, rpc string, check_content_type bool) bool {
	if check_content_type {
		if utils.BytesToString(r.Request.Header.ContentType()) != fmt.Sprintf("application/x-git-%s-request", rpc) {
			return false
		}
	}

	if !(rpc == "upload-pack" || rpc == "receive-pack") {
		return false
	}
	if rpc == "receive-pack" {
		return this.config.ReceivePack
	}
	if rpc == "upload-pack" {
		return this.config.UploadPack
	}

	return this.getConfigSetting(rpc, dir)
}

func (this *GitHttp) getConfigSetting(service_name string, dir string) bool {
	service_name = strings.Replace(service_name, "-", "", -1)
	setting := this.getGitConfig("http."+service_name, dir)

	if service_name == "uploadpack" {
		return setting != "false"
	}

	return setting == "true"
}

func (this *GitHttp) getGitConfig(config_name string, dir string) string {
	args := []string{"config", config_name}
	out := utils.BytesToString(this.gitCommand(dir, args...))
	return out[0 : len(out)-1]
}

func (this *GitHttp) updateServerInfo(dir string) []byte {
	args := []string{"update-server-info"}
	return this.gitCommand(dir, args...)
}

func (this *GitHttp) gitCommand(dir string, args ...string) []byte {
	command := exec.Command(this.config.GitBinPath, args...)
	command.Dir = dir
	out, err := command.Output()

	if err != nil {
		log.Print(err)
	}

	return out
}

// HTTP error response handling functions

func (this *GitHttp) renderMethodNotAllowed(r *fasthttp.RequestCtx) {
	if r.Request.Header.IsHTTP11() {
		r.SetStatusCode(http.StatusMethodNotAllowed)
		r.Write([]byte("Method Not Allowed"))
	} else {
		r.SetStatusCode(http.StatusBadRequest)
		r.Write([]byte("Bad Request"))
	}
}

func (this *GitHttp) renderNotFound(r *fasthttp.RequestCtx) {
	r.SetStatusCode(http.StatusNotFound)
	r.Write([]byte("Not Found"))
}

func (this *GitHttp) renderNoAccess(r *fasthttp.RequestCtx) {
	r.SetStatusCode(http.StatusForbidden)
	r.Write([]byte("Forbidden"))
}

func (this *GitHttp) renderUnauthorized(r *fasthttp.RequestCtx) {
	r.SetStatusCode(http.StatusUnauthorized)
	r.Write([]byte("Unauthorized"))
}

// Packet-line handling function

func (this *GitHttp) packetFlush() []byte {
	return []byte("0000")
}

func (this *GitHttp) packetWrite(str string) []byte {
	s := strconv.FormatInt(int64(len(str)+4), 16)

	if len(s)%4 != 0 {
		s = strings.Repeat("0", 4-len(s)%4) + s
	}

	return []byte(s + str)
}

// Header writing functions

func (this *GitHttp) hdrNocache(w *fasthttp.RequestCtx) {
	w.Response.Header.Set("Expires", "Fri, 01 Jan 1980 00:00:00 GMT")
	w.Response.Header.Set("Pragma", "no-cache")
	w.Response.Header.Set("Cache-Control", "no-cache, max-age=0, must-revalidate")
}

func (this *GitHttp) hdrCacheForever(w *fasthttp.RequestCtx) {
	now := time.Now().Unix()
	expires := now + 31536000
	w.Response.Header.Set("Date", fmt.Sprintf("%d", now))
	w.Response.Header.Set("Expires", fmt.Sprintf("%d", expires))
	w.Response.Header.Set("Cache-Control", "public, max-age=31536000")
}

func (this *GitHttp) Start() {
	err := fasthttp.ListenAndServe(this.config.ListenAddr, this.requestHandler)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func (this *GitHttp) RegisterVerify(v Verification) {
	this.verify = v
}

func New(address string, projectBaseDir string) *GitHttp {
	githttp := &GitHttp{
		config: &Config{
			ListenAddr:  address,
			ProjectBase: projectBaseDir,
			GitBinPath:  "/usr/bin/git",
			UploadPack:  true,
			ReceivePack: true,
		},
	}

	services := map[string]Service{
		"(.*?)/git-upload-pack$":                       Service{"POST", githttp.serviceRpc, "upload-pack"},
		"(.*?)/git-receive-pack$":                      Service{"POST", githttp.serviceRpc, "receive-pack"},
		"(.*?)/info/refs$":                             Service{"GET", githttp.getInfoRefs, ""},
		"(.*?)/HEAD$":                                  Service{"GET", githttp.getTextFile, ""},
		"(.*?)/objects/info/alternates$":               Service{"GET", githttp.getTextFile, ""},
		"(.*?)/objects/info/http-alternates$":          Service{"GET", githttp.getTextFile, ""},
		"(.*?)/objects/info/packs$":                    Service{"GET", githttp.getInfoPacks, ""},
		"(.*?)/objects/info/[^/]*$":                    Service{"GET", githttp.getTextFile, ""},
		"(.*?)/objects/[0-9a-f]{2}/[0-9a-f]{38}$":      Service{"GET", githttp.getLooseObject, ""},
		"(.*?)/objects/pack/pack-[0-9a-f]{40}\\.pack$": Service{"GET", githttp.getPackFile, ""},
		"(.*?)/objects/pack/pack-[0-9a-f]{40}\\.idx$":  Service{"GET", githttp.getIdxFile, ""},
	}
	githttp.services = services
	return githttp
}

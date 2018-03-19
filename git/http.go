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
	"compress/gzip"
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

	"github.com/molisoft/v2git/cfg"
	"github.com/molisoft/v2git/utils"
)

type Service struct {
	Method  string
	Handler func(HandlerReq)
	Rpc     string
}

type HandlerReq struct {
	w    http.ResponseWriter
	r    *http.Request
	Rpc  string
	Dir  string
	File string
}

type GitHttp struct {
	GitBase
	config   *Config
	services map[string]Service
	verify   Verification
}

// Request handling function

func (this *GitHttp) requestHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s %s %s", r.RemoteAddr, r.Method, r.URL.Path, r.Proto)
	if !this.verification(w, r) {
		this.renderUnauthorized(w)
		return
	}

	for match, service := range this.services {
		re, err := regexp.Compile(match)
		if err != nil {
			log.Print(err)
		}

		if m := re.FindStringSubmatch(r.URL.Path); m != nil {
			if service.Method != r.Method {
				this.renderMethodNotAllowed(w, r)
				return
			}

			rpc := service.Rpc
			file := strings.Replace(r.URL.Path, m[1]+"/", "", 1)
			dir, err := this.FullPath(m[1])
			if err != nil {
				log.Println("Full path error ", err)
				this.renderNotFound(w)
				return
			}

			hr := HandlerReq{w, r, rpc, dir, file}
			service.Handler(hr)
			return
		}
	}
	this.renderNotFound(w)
	return

}

// 验证账号密码

func (this *GitHttp) verification(w http.ResponseWriter, r *http.Request) bool {
	if this.verify == nil {
		return true
	}
	authField := r.Header.Get("Authorization")
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
					gitPath := r.URL.Path

					authRequest := &cfg.AuthRequest{
						AuthURL:  this.config.AuthUrl,
						Path:     gitPath,
						Username: username,
						Password: password,
					}

					return this.verify(authRequest)
				}
			}
		}
	}
	w.Header().Set("WWW-Authenticate", fmt.Sprintf("Basic realm=\"%s\"", Realm))
	log.Println("Unauthorized!")
	return false
}

// Actual command handling functions

func (this *GitHttp) serviceRpc(hr HandlerReq) {
	w, r, rpc, dir := hr.w, hr.r, hr.Rpc, hr.Dir
	access := this.hasAccess(r, dir, rpc, true)

	if access == false {
		this.renderNoAccess(w)
		return
	}

	w.Header().Set("Content-Type", fmt.Sprintf("application/x-git-%s-result", rpc))
	w.WriteHeader(http.StatusOK)

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

	var reader io.ReadCloser
	switch r.Header.Get("Content-Encoding") {
	case "gzip":
		reader, _ = gzip.NewReader(r.Body)
		defer reader.Close()
	default:
		reader = r.Body
	}
	io.Copy(in, reader)
	in.Close()
	io.Copy(w, stdout)
	cmd.Wait()
}

func (this *GitHttp) getInfoRefs(hr HandlerReq) {
	w, r, dir := hr.w, hr.r, hr.Dir
	service_name := this.getServiceType(r)
	access := this.hasAccess(r, dir, service_name, false)

	if access {
		args := []string{service_name, "--stateless-rpc", "--advertise-refs", "."}
		fmt.Println(args)
		refs := this.gitCommand(dir, args...)

		this.hdrNocache(w)
		w.Header().Set("Content-Type", fmt.Sprintf("application/x-git-%s-advertisement", service_name))
		w.WriteHeader(http.StatusOK)
		w.Write(this.packetWrite("# service=git-" + service_name + "\n"))
		w.Write(this.packetFlush())
		w.Write(refs)
	} else {
		this.updateServerInfo(dir)
		this.hdrNocache(w)
		this.sendFile("text/plain; charset=utf-8", hr)
	}
}

func (this *GitHttp) getInfoPacks(hr HandlerReq) {
	this.hdrCacheForever(hr.w)
	this.sendFile("text/plain; charset=utf-8", hr)
}

func (this *GitHttp) getLooseObject(hr HandlerReq) {
	this.hdrCacheForever(hr.w)
	this.sendFile("application/x-git-loose-object", hr)
}

func (this *GitHttp) getPackFile(hr HandlerReq) {
	this.hdrCacheForever(hr.w)
	this.sendFile("application/x-git-packed-objects", hr)
}

func (this *GitHttp) getIdxFile(hr HandlerReq) {
	this.hdrCacheForever(hr.w)
	this.sendFile("application/x-git-packed-objects-toc", hr)
}

func (this *GitHttp) getTextFile(hr HandlerReq) {
	this.hdrNocache(hr.w)
	this.sendFile("text/plain", hr)
}

// Logic helping functions

func (this *GitHttp) sendFile(content_type string, hr HandlerReq) {
	w, r := hr.w, hr.r
	req_file := path.Join(hr.Dir, hr.File)

	f, err := os.Stat(req_file)
	if os.IsNotExist(err) {
		this.renderNotFound(w)
		return
	}
	w.Header().Set("Content-Type", content_type)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", f.Size()))
	w.Header().Set("Last-Modified", f.ModTime().Format(http.TimeFormat))
	http.ServeFile(w, r, req_file)
}

func (this *GitHttp) getServiceType(r *http.Request) string {
	service_type := r.FormValue("service")

	if s := strings.HasPrefix(service_type, "git-"); !s {
		return ""
	}

	return strings.Replace(service_type, "git-", "", 1)
}

func (this *GitHttp) hasAccess(r *http.Request, dir string, rpc string, check_content_type bool) bool {
	if check_content_type {
		if r.Header.Get("Content-Type") != fmt.Sprintf("application/x-git-%s-request", rpc) {
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

func (this *GitHttp) renderMethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	if r.Proto == "HTTP/1.1" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Method Not Allowed"))
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad Request"))
	}
}

func (this *GitHttp) renderNotFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Not Found"))
}

func (this *GitHttp) renderNoAccess(w http.ResponseWriter) {
	w.WriteHeader(http.StatusForbidden)
	w.Write([]byte("Forbidden"))
}

func (this *GitHttp) renderUnauthorized(w http.ResponseWriter) {
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte("Unauthorized"))
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

func (this *GitHttp) hdrNocache(w http.ResponseWriter) {
	w.Header().Set("Expires", "Fri, 01 Jan 1980 00:00:00 GMT")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Cache-Control", "no-cache, max-age=0, must-revalidate")
}

func (this *GitHttp) hdrCacheForever(w http.ResponseWriter) {
	now := time.Now().Unix()
	expires := now + 31536000
	w.Header().Set("Date", fmt.Sprintf("%d", now))
	w.Header().Set("Expires", fmt.Sprintf("%d", expires))
	w.Header().Set("Cache-Control", "public, max-age=31536000")
}

func (this *GitHttp) Start() {
	http.HandleFunc("/", this.requestHandler)
	err := http.ListenAndServe(this.config.ListenAddr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func (this *GitHttp) RegisterVerify(v Verification) {
	this.verify = v
}

func NewHTTP(address string, cfg *cfg.Config) *GitHttp {
	githttp := &GitHttp{
		config: &Config{
			ListenAddr:  address,
			RepoRoot:    cfg.RepoDir,
			GitBinPath:  cfg.GitBinPath,
			AuthUrl:     cfg.AuthUrl,
			UploadPack:  true,
			ReceivePack: true,
		},
		GitBase: GitBase{
			RootPath: cfg.RepoDir,
		},
	}

	githttp.services = map[string]Service{
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
	return githttp
}

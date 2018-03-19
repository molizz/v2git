package git

import (
	"io"
	"log"
	"os/exec"
	"time"

	"github.com/gliderlabs/ssh"
	"github.com/molisoft/v2git/cfg"
	gossh "golang.org/x/crypto/ssh"
)

type GitSSH struct {
	config          *Config
	verify          Verification      // 验证回调
	allowedCommends map[string]string // 允许的命令
	allowedUser     string            // 允许的用户
}

var (
	errBadUser = []byte("Bad user\n")
	//errRepoNotFound = []byte("Repo not found\n")
)

func (this *GitSSH) isGitUser(currentUser string) bool {
	return this.allowedUser == currentUser
}

func (this *GitSSH) Handler(session ssh.Session) {
	var err error
	commands := session.Command()
	log.Println("Request command: ", commands)

	// 验证授权
	if this.verify != nil {
		fingerprint := gossh.FingerprintLegacyMD5(session.PublicKey())
		gitPath := commands[1]
		authRequest := &cfg.AuthRequest{
			Fingerprint: fingerprint,
			Path:        gitPath,
		}
		if !this.verify(authRequest) {
			return
		}
	}

	// 获取命令，并判断命令是否可用
	service, ok := this.allowedCommends[commands[0]]
	if !ok {
		return
	}

	// URL.path转换成路径
	logicPath, err := UrlToNamespace(commands[1])
	if err != nil {
		log.Println("Path to namespace error ", err)
		return
	}
	// 连接成绝对物理路径
	dir := this.repoFullPath(logicPath)

	// 开始执行git命令
	args := []string{service, dir}
	cmd := exec.Command(this.config.GitBinPath, args...)
	cmd.Dir = dir
	in, err := cmd.StdinPipe()
	if err != nil {
		log.Println("get the stdin pipe error ", err)
		return
	}

	out, err := cmd.StdoutPipe()
	if err != nil {
		log.Println("Get the stdout pipe error ", err)
		return
	}
	err = cmd.Start()
	if err != nil {
		log.Println("Start the cmd error ", err)
		return
	}

	go func() {
		io.Copy(in, session)
		in.Close()
		log.Println("Session id: " + session.RemoteAddr().String() + " -> session to in ")
	}()

	go func() {
		io.Copy(session, out)
		log.Println("Session id: " + session.RemoteAddr().String() + " -> out to session")
	}()
	cmd.Wait()
	log.Println("Session id: " + session.RemoteAddr().String() + " -> signout ")
}

func (this *GitSSH) repoFullPath(requestPath string) string {
	gitDir := this.config.RepoRoot + "/" + requestPath
	log.Println("Git full path to the repo: ", gitDir)
	return gitDir
}

func (this *GitSSH) Start() {
	err := ssh.ListenAndServe(
		this.config.ListenAddr,
		this.Handler,
		ssh.PublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
			log.Println("New session ", ctx.RemoteAddr())
			if !this.isGitUser(ctx.User()) {
				log.Println("The user " + ctx.User() + " is not " + this.allowedUser)
				return false
			}
			// 无论如何先允许ssh连入
			return true
		}),
		ssh.HostKeyFile(this.config.PrivateKeyPath),
		func(server *ssh.Server) error {
			server.IdleTimeout = 60 * time.Second
			server.MaxTimeout = 60 * time.Second
			server.Version = Realm
			return nil
		},
	)

	if err != nil {
		log.Fatalf("Failed SSH: %s", err.Error())
	}
}

func (this *GitSSH) RegisterVerify(v Verification) {
	this.verify = v
}

/*
	addr = ":22"
*/
func NewSSH(addr string, cfg *cfg.Config) *GitSSH {
	gitssh := &GitSSH{
		config: &Config{
			RepoRoot:       cfg.RepoDir,
			AuthUrl:        cfg.AuthUrl,
			PrivateKeyPath: cfg.PrivateKeyPath,
			GitBinPath:     cfg.GitBinPath,
			ListenAddr:     addr,
			UploadPack:     true,
			ReceivePack:    true,
		},
		allowedUser: cfg.GitUser,
		allowedCommends: map[string]string{
			"git-receive-pack":   "receive-pack",
			"git-upload-pack":    "upload-pack",
			"git-upload-archive": "upload-archive",
		},
	}
	return gitssh
}

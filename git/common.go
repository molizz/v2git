package git

import "github.com/molisoft/v2git/cfg"

type Verification func(*cfg.AuthRequest) bool

type Giter interface {
	Start()
	RegisterVerify(Verification)
}

type Config struct {
	ListenAddr     string
	RepoRoot       string
	GitBinPath     string
	AuthUrl        string
	PrivateKeyPath string
	UploadPack     bool
	ReceivePack    bool
}

var (
	Realm = "V2Git"
)

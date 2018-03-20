package cfg

import (
	"github.com/jinzhu/configor"
)

type Config struct {
	HttpPort       uint   `json:"http_port"`
	SshPort        uint   `json:"ssh_port"`
	RepoDir        string `json:"repo_dir"`
	PrivateKeyPath string `json:"private_key_path"`
	AuthUrl        string `json:"auth_url"`
	GitBinPath     string `json:"git_bin_path"`
	GitUser        string `json:"git_user"`
	ApiPassword    string `json:"api_password"`
}

func New(configPath string) (*Config, error) {
	cfg := &Config{}
	err := configor.Load(cfg, configPath)
	return cfg, err
}

type AuthRequest struct {
	Path        string `json:"path"`        // xxx/xxx.git
	AuthURL     string `json:"auth_url"`    // 授权url
	Fingerprint string `json:"fingerprint"` // ssh验证的公有key指纹
	Username    string `json:"username"`    // http验证的账号
	Password    string `json:"password"`    // http验证的密码
}

type AuthResponse struct {
	Result int `json:"result"`
}

func (this *AuthResponse) IsOK() bool {
	return this.Result == 0
}

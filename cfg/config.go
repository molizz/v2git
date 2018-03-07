package cfg

import (
	"github.com/jinzhu/configor"
)

type Config struct {
	Port    uint   `json:"port"`
	RepoDir string `json:"repo_dir"`
	AuthUrl string `json:"auth_url"`
}

func New(configPath string) (*Config, error) {
	cfg := &Config{}
	err := configor.Load(cfg, configPath)
	return cfg, err
}

type AuthRequest struct {
	AuthURL  string `json:"auth_url"`
	Path     string `json:"path"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Result int `json:"result"`
}

func (this *AuthResponse) IsOK() bool {
	return this.Result == 0
}

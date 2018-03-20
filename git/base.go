package git

import (
	"os"
	"path"

	"github.com/molisoft/v2git/utils"
)

type GitBase struct {
	RootPath string
}

//
//	根据请求的path（/xxx/xx.git） 返回完整的物理路径
//
func (this *GitBase) FullPath(requestPath string) (string, error) {

	dir, err := utils.UrlToDirPath(requestPath)
	if err != nil {
		return "", err
	}
	dir += ".git"

	f := path.Join(this.RootPath, dir)
	if _, err := os.Stat(f); os.IsNotExist(err) {
		return "", err
	}
	return f, nil
}

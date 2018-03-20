package repo

import (
	"errors"
	"os"
	"path"

	"github.com/molisoft/v2git/utils"
	"gopkg.in/libgit2/git2go.v25"
)

var RootDir string // 仓库存储根目录

var (
	ErrRepoExists = errors.New("Repo is exists.")
)

type Repository struct {
	RepoPath string // 仓库目录
	repo     *git.Repository
}

func New(namespace string) (*Repository, error) {
	repoPath, err := utils.UrlToDirPath(namespace)
	if err != nil {
		return nil, err
	}
	repo := &Repository{
		RepoPath: path.Join(RootDir, repoPath+".git"),
	}
	err = repo.openRepository()
	if err != nil {
		return nil, err
	}
	return repo, nil
}

// create bare repository
//
func (this *Repository) CreateProject() error {
	if this.IsExists() {
		return ErrRepoExists
	}

	var err error
	this.repo, err = git.InitRepository(this.RepoPath, true)
	if err != nil {
		return err
	}
	return nil
}

func (this *Repository) IsExists() bool {
	_, err := os.Stat(this.RepoPath)
	return !os.IsNotExist(err)
}

func (this *Repository) openRepository() (err error) {
	if this.IsExists() {
		this.repo, err = git.OpenRepository(this.RepoPath)
	}
	return
}

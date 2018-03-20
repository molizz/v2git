package repo

import (
	"errors"
	"os"
	"path"

	"github.com/molisoft/v2git/utils"
	"gopkg.in/src-d/go-billy.v4/osfs"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/storage/filesystem"
)

var RootDir string // 仓库存储根目录

var (
	ErrRepoExists = errors.New("Repo is exists.")
)

type Repository struct {
	RepoPath     string // 仓库目录
	repo         *git.Repository
	repo_storage *filesystem.Storage
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
	fs, err := this.Storage()
	this.repo, err = git.Init(fs, nil)

	if err != nil {
		return err
	}
	return nil
}

func (this *Repository) DeleteProject() error {
	if this.repo != nil {
		//this.repo.Storer
	}
	return os.RemoveAll(this.RepoPath)
}

func (this *Repository) IsExists() bool {
	_, err := os.Stat(this.RepoPath)
	return !os.IsNotExist(err)
}

func (this *Repository) openRepository() (err error) {
	if this.IsExists() {
		this.repo_storage, err = this.Storage()
		this.repo, err = git.Open(this.repo_storage, nil)
	}
	return err
}

func (this *Repository) Storage() (*filesystem.Storage, error) {
	filesystem.NewStorage(osfs.New(this.RepoPath))
}

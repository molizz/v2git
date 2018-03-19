package git

import (
	"errors"
	"net/url"
	"strings"
)

var (
	errBadPath    = errors.New("Bad path")
	errBadRequest = errors.New("Bad request")
)

/*
	http://xxx.com/moli/abc.git to moli/abc
*/
func UrlToNamespace(rawurl string) (path string, err error) {
	uri, err := url.Parse(rawurl)
	if err != nil {
		return
	}

	path = uri.Path
	path = strings.TrimLeft(path, "/")
	path = strings.TrimSuffix(path, ".git")
	return
}

/*
	moli/abc to ab/mo/moli/abc

	应防止用户发送 ../.. 等
*/
func NamespaceToDirPath(namespace string) (dir string, err error) {
	paths := strings.SplitN(namespace, "/", 2)
	if len(paths) != 2 {
		err = errBadRequest
		return
	}

	var dir1, dir2, dir3, dir4 = paths[1], paths[0], paths[0], paths[1]

	if dir1 == ".." || dir2 == ".." {
		err = errBadPath
		return
	}

	if len(dir1) < 2 && len(dir1) > 0 {
		dir1 = "0" + dir1 // 如果路径只有1位时补0
	}
	if len(dir2) < 2 && len(dir2) > 0 {
		dir2 = "0" + dir2
	}

	dirs := []string{dir1[:2], dir2[:2], dir3, dir4}
	dir = strings.Join(dirs, "/")
	return
}

/*
	http://xxx.com/moli/abc.git to ab/mo/moli/abc
*/
func UrlToDirPath(url string) (dir string, err error) {
	namespace, err := UrlToNamespace(url)
	if err != nil {
		return
	}
	dir, err = NamespaceToDirPath(namespace)
	return
}

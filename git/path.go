package git

import (
	"errors"
	"strings"

	"github.com/molisoft/v2git/utils"
	"github.com/valyala/fasthttp"
)

/*
	http://xxx.com/moli/abc.git to moli/abc
*/
func UrlToNamespace(url string) string {
	var uri fasthttp.URI
	uri.Update(url)

	path := utils.BytesToString(uri.Path())
	path = strings.TrimLeft(path, "/")
	path = strings.TrimSuffix(path, ".git")
	return path
}

/*
	moli/abc to ab/mo/moli/abc
*/
func NamespaceToDirPath(namespace string) (dir string, err error) {
	paths := strings.SplitN(namespace, "/", 2)
	if len(paths) != 2 {
		err = errors.New("path exception.")
		return
	}

	dirs := []string{paths[1][:2], paths[0][:2], paths[0], paths[1]}
	dir = strings.Join(dirs, "/")
	return
}

/*
	http://xxx.com/moli/abc.git to ab/mo/moli/abc
*/
func UrlToDirPath(url string) (dir string, err error) {
	namespace := UrlToNamespace(url)
	dir, err = NamespaceToDirPath(namespace)
	return
}

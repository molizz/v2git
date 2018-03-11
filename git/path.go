package git

import (
	"net/url"
	"strings"
)

/*
	http://xxx.com/moli/abc.git to moli/abc.git
*/
func UrlToNamespace(rawurl string) (path string, err error) {
	uri, err := url.Parse(rawurl)
	if err != nil {
		return
	}

	path = uri.Path
	path = strings.TrimPrefix(path, "/")
	return
}

package git

import "testing"

func TestUriToPath(t *testing.T) {
	url1 := "https://v2geek.com/moli/xxx.git"
	if path, err := UrlToNamespace(url1); err != nil && path != "moli/xxx.git" {
		t.Error(UrlToNamespace(url1))
	}

	url2 := "/mo-li/x-x.git"
	if path, err := UrlToNamespace(url2); err != nil && path != "mo-li/x-x.git" {
		t.Error(UrlToNamespace(url1))
	}
}

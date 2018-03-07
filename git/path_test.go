package git

import "testing"

func TestUriToPath(t *testing.T) {
	url1 := "https://v2geek.com/moli/xxx.git"
	if UrlToNamespace(url1) != "moli/xxx" {
		t.Error(UrlToNamespace(url1))
	}

	url2 := "/mo-li/x-x.git"
	if UrlToNamespace(url2) != "mo-li/x-x" {
		t.Error(UrlToNamespace(url1))
	}
}

func TestPathToDirPath(t *testing.T) {
	url := "https://v2geek.com/moli/xxx.git"
	ns := UrlToNamespace(url)
	dir, err := NamespaceToDirPath(ns)
	if err != nil {
		t.Error(err)
	}
	if dir != "xx/mo/moli/xxx" {
		t.Error("NamespaceToDirPath error！", dir)
	}
}

func TestUrlToDirPath(t *testing.T) {
	url := "https://v2geek.com/moli/xxx.git"
	dir, err := UrlToDirPath(url)
	if err != nil {
		t.Error(err)
	}

	if dir != "xx/mo/moli/xxx" {
		t.Error("UrlToDirPath error！", dir)
	}
}

package utils

import (
	"bytes"
	"testing"
)

func TestBytesToString(t *testing.T) {
	a := []byte("hello")
	if BytesToString(a) != "hello" {
		t.Fail()
	}
}

func TestStringToBytes(t *testing.T) {
	a := "hello"
	if bytes.Compare(StringToBytes(a), []byte("hello")) != 0 {
		t.Fail()
	}
}

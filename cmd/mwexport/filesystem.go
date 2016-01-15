package main

import (
	"io/ioutil"
	"os"
)

type fileSystem interface {
	WriteFile(filename string, data []byte, perm os.FileMode) error
}

type localFileSystem struct{}

func (localFileSystem) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return ioutil.WriteFile(filename, data, perm)
}

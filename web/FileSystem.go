package web

import (
	"errors"
	"net/http"
	"os"
	"path"
	"strings"
)

// blocks directory listing and accessing content outside the document base (root)
type OnlyFilesFS struct {
	Fs http.FileSystem
}

func (this OnlyFilesFS) Open(name string) (http.File, error) {
	path := path.Clean(name)
	// avoids access outside of root
	if strings.HasPrefix(path, "..") {
		return nil, errors.New(name + " is not a valid path.")
	}
	//logger.Debugf("opening name: %s", name)
	f, err := this.Fs.Open(name)
	if err != nil {
		return nil, err
	}
	return NeuteredReadirFile{f}, nil
}

type NeuteredReadirFile struct {
	http.File
}

func (f NeuteredReadirFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, nil
}

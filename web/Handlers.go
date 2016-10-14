package web

import (
	"bytes"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var _ http.ResponseWriter = &BufferedResponse{}

type BufferedResponse struct {
	Body   *bytes.Buffer
	header http.Header
	Code   int
}

func NewBufferedResponse() *BufferedResponse {
	return &BufferedResponse{
		Body:   new(bytes.Buffer),
		header: make(http.Header),
	}
}

func (this *BufferedResponse) Header() http.Header {
	return this.header
}

func (this *BufferedResponse) Write(data []byte) (int, error) {
	var n, err = this.Body.Write(data)
	if err == nil && this.Code == 0 {
		this.Code = http.StatusOK
	}
	return n, err
}

// WriteBefore writes data at the beginning of the buffer.
// It does this by copying the original data, truncating the buffer, then writing
// what we want and then write the original data.
func (this *BufferedResponse) WriteBefore(data []byte) error {
	var body = this.Body.Bytes()
	var clone = make([]byte, len(body))
	copy(clone, body)

	this.Body.Reset()
	if _, err := this.Body.Write(data); err != nil {
		return err
	}
	if len(clone) > 0 {
		if _, err := this.Body.Write(clone); err != nil {
			return err
		}
	}
	if this.Code == 0 {
		this.Code = http.StatusOK
	}
	return nil
}

func (this *BufferedResponse) WriteHeader(code int) {
	this.Code = code
}

func (this *BufferedResponse) Flush(w http.ResponseWriter) {
	// we copy the original headers first
	for k, v := range this.Header() {
		w.Header()[k] = v
	}

	// status code
	w.WriteHeader(this.Code)

	// then write out the original body
	w.Write(this.Body.Bytes())
}

type resourcesHandler struct {
	// root directory
	dir string
}

const gzipExt = ".gz"

// ResourcesHandler creates a handler that serves files from a folder.
// If the broowser accepts gzip encoding, it will first try to serve the gzipped version of the file.
// If absent it will serve the non-gizzepd file.
// The gzip file name is the original file nane concatenated with '.gz'.
// Ex: index.html -> index.html.gz
func ResourcesHandler(dir string) http.Handler {
	return resourcesHandler{dir}
}

func (fh resourcesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fname := r.URL.Path
	if fname == "/" {
		fname = "/index.html"
	}

	fname = filepath.Join(fh.dir, fname)

	var useGzip bool
	for _, v := range r.Header["Accept-Encoding"] {
		if strings.Contains(v, "gzip") {
			useGzip = true
			break
		}
	}
	// Check if the file exists. If not, attempt to serve a non-gzipped version.
	var f os.FileInfo
	var err error
	if useGzip {
		f, err = os.Stat(fname + gzipExt)
	}
	if f == nil {
		useGzip = false
		// Attempt to serve the non-gzipped version
		f, err = os.Stat(fname)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Resource not found!"))
			return
		}
	}

	// Don't show directory listings.
	if f.IsDir() {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Directory listing is not allowed!"))
		return
	}

	if useGzip {
		w.Header().Set("Content-Encoding", "gzip")
		ctype := mime.TypeByExtension(filepath.Ext(fname))
		if ctype != "" {
			w.Header().Set("Content-Type", ctype)
		}
		fname += gzipExt
	}

	// http.ServeFile sets Last-Modified headers based on modtime for us.
	http.ServeFile(w, r, fname)
}

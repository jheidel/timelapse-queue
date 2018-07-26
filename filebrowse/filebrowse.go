package filebrowse

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type FileBrowser struct {
	// Root is the base of the file system to serve up
	Root string
}

type Directory struct {
	Name string
	Path string
}

type Timelapse struct {
}

type Response struct {
	Parents    []*Directory
	Dirs       []*Directory
	Timelapses []*Timelapse
}

func (f *FileBrowser) listPath(p string) (*Response, error) {
	root, err := filepath.EvalSymlinks(f.Root)
	if err != nil {
		return nil, err
	}
	b, err := filepath.EvalSymlinks(filepath.Join(root, p))
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(b, root) {
		return nil, errors.New("permission denied, not in root")
	}

	files, err := ioutil.ReadDir(b)
	if err != nil {
		return nil, err
	}

	r := &Response{}

	// Generate list of directories
	for _, finfo := range files {
		rel, err := filepath.Rel(root, filepath.Join(b, finfo.Name()))
		if err != nil {
			return nil, err
		}
		if finfo.IsDir() {
			d := &Directory{
				Name: finfo.Name(),
				Path: rel,
			}
			r.Dirs = append(r.Dirs, d)
		}
	}

	// Generates list of parents
	rel, err := filepath.Rel(root, b)
	if err != nil {
		return nil, err
	}
	pl := strings.Split(rel, string(os.PathSeparator))
	r.Parents = append(r.Parents, &Directory{
		Name: "[top]",
		Path: ".",
	})
	for i := range pl {
		d := &Directory{
			Name: pl[i],
			Path: filepath.Join(pl[:i+1]...),
		}
		if d.Path != "." {
			r.Parents = append(r.Parents, d)
		}
	}

	return r, nil
}

func (f *FileBrowser) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	p := r.Form.Get("path")
	response, err := f.listPath(p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	js, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

package filebrowse

import (
	"encoding/json"
	"net/http"
)

type FileBrowser struct {
	// Root is the base of the file system to serve up
	Root string
}

type Timelapse struct {
}

type Response struct {
	Parents    []string
	DirNames   []string
	Timelapses []*Timelapse
}

func (f *FileBrowser) listPath(path string) (*Response, error) {
	return nil, nil
}

func (f *FileBrowser) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	p := r.Form.Get("path")
	if p == "" {
		http.Error(w, "missing path", http.StatusBadRequest)
		return
	}
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

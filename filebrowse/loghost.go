package filebrowse

import (
	"io"
	"net/http"
	"os"
)

type LogHost struct {
	Browser *FileBrowser
}

func (h *LogHost) writeLog(path, name string, w http.ResponseWriter) error {
	t, err := h.Browser.GetTimelapse(path)
	if err != nil {
		return err
	}

	txt, err := os.Open(t.GetOutputFullPath(name))
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	_, err = io.Copy(w, txt)
	if err != nil {
		return err
	}

	return nil
}

func (h *LogHost) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	path := r.Form.Get("path")
	name := r.Form.Get("name")
	if err := h.writeLog(path, name, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

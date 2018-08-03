package filebrowse

import (
	"io"
	"net/http"
	"os"
)

type LogHost struct {
	Browser *FileBrowser
}

func (h *LogHost) writeLog(rel string, w http.ResponseWriter) error {
	path, err := h.Browser.GetFullPath(rel)
	if err != nil {
		return err
	}

	txt, err := os.Open(path)
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
	rel := r.Form.Get("path")
	if err := h.writeLog(rel, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

package engine

import (
	"net/http"
	"net/url"
	"strconv"

	"timelapse-queue/filebrowse"
)

type TestServer struct {
	Browser *filebrowse.FileBrowser
	Queue   *JobQueue
}

func parseBounds(values url.Values) (*Bounds, error) {
	var err error
	b := &Bounds{}
	b.X, err = strconv.Atoi(values.Get("x"))
	if err != nil {
		return nil, err
	}
	b.Y, err = strconv.Atoi(values.Get("y"))
	if err != nil {
		return nil, err
	}
	b.Width, err = strconv.Atoi(values.Get("width"))
	if err != nil {
		return nil, err
	}
	b.Height, err = strconv.Atoi(values.Get("height"))
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (s *TestServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Requires POST", http.StatusBadRequest)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	rel := r.Form.Get("path")
	t, err := s.Browser.GetTimelapse(rel)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	b, err := parseBounds(r.Form)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	config := &configFake{
		Bounds: b,
	}

	if err := config.Validate(t); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.Queue.AddJob(config, t)
}

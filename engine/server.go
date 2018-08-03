package engine

import (
	"net/http"

	"timelapse-queue/filebrowse"
)

type TestServer struct {
	Browser *filebrowse.FileBrowser
	Queue   *JobQueue
}

func (s *TestServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	config := &configFake{}

	s.Queue.AddJob(config, t)
	w.Write([]byte("done"))
}

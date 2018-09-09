package engine

import (
	"encoding/json"
	"net/http"

	"github.com/davecgh/go-spew/spew"
	log "github.com/sirupsen/logrus"
	"timelapse-queue/filebrowse"
)

type TestServer struct {
	Browser *filebrowse.FileBrowser
	Queue   *JobQueue
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

	config := &baseConfig{}
	if err := json.Unmarshal([]byte(r.Form.Get("request")), config); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Infof("Received conversion request: %+v", spew.Sdump(config))

	t, err := s.Browser.GetTimelapse(config.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err := config.Validate(r.Context(), t); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.Queue.AddJob(config, t)
}

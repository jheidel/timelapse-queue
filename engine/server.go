package engine

import (
	"context"
	"net/http"
	"time"

	"timelapse-queue/filebrowse"

	log "github.com/sirupsen/logrus"
)

type TestServer struct {
	Browser *filebrowse.FileBrowser
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

	pc := make(chan int)
	go func() {
		for p := range pc {
			log.Infof("Chan progress %d", p)
		}

		log.Infof("Chan exit.")
	}()

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Minute)
	defer cancel()

	err = Convert(ctx, config, t, pc)
	if err != nil {
		log.Errorf("Convert returned error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Infof("Conversion succeeded.")
	w.Write([]byte("done"))
}

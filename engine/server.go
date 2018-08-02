package engine

import (
	"context"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

type TestServer struct {
}

func (s *TestServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

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

	err := Convert(ctx, config, nil, pc)
	if err != nil {
		log.Errorf("Convert returned error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Infof("Conversion succeeded.")
	w.Write([]byte("done"))
}

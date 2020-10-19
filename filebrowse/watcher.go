package filebrowse

import (
	"io/ioutil"
	"time"

	log "github.com/sirupsen/logrus"
)

func check(root string) {
	files, err := ioutil.ReadDir(root)
	if err != nil {
		log.Fatalf("Failed to read files in %q: %v", root, err)
	}
	if len(files) == 0 {
		log.Fatalf("Found no files in %q", root)
	}
}

func WatchMountHealthy(root string) {
	check(root)
	go func() {
		t := time.NewTicker(time.Minute)
		for _ = range t.C {
			check(root)
		}
	}()
}

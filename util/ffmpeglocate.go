package util

import (
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
)

// LocateFFmpeg finds the location of the ffmpeg binary, looking in common locations.
func LocateFFmpeg() (string, error) {
	// Check environment.
	if p := os.Getenv("FFMPEG"); p != "" {
		return p, nil
	}

	// Check PATH.
	p, err := exec.LookPath("ffmpeg")
	if err != nil {
		return "", err
	}
	return p, nil
}

// LocateFFmpegOrDie is equivalent to LocateFFmpeg, but will panic if not found.
func LocateFFmpegOrDie() string {
	p, err := LocateFFmpeg()
	if err != nil {
		log.Fatalf("Unable to locate ffmpeg: %v", err)
	}
	return p
}

package engine

import (
	"timelapse-queue/filebrowse"
)

type Config interface {
	GetArgs(timelapse *filebrowse.Timelapse) []string
}

type configFake struct {
}

func (f *configFake) GetArgs(t *filebrowse.Timelapse) []string {
	return []string{
		"-r", "60",
		"-start_number", "21015",
		"-i", "/home/jeff/timelapse/G%07d.JPG",
		"-c:v", "libx264",
		"-preset", "slow",
		"-crf", "17",
		"-s", "1920x1080",
		"/home/jeff/timelapse/1080p-test.mp4",
	}
}

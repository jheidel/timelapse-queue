package engine

import (
	"fmt"

	"timelapse-queue/filebrowse"
)

type Config interface {
	GetArgs(timelapse *filebrowse.Timelapse) []string
	GetDebugFullPath(timelapse *filebrowse.Timelapse) string
	GetDebugPath(timelapse *filebrowse.Timelapse) string
}

type configFake struct {
}

// TODO: rest of this class. Some things like progress implementation are
// ffmpeg internal and should be pushed into ffmpeg.go.

func (f *configFake) GetArgs(t *filebrowse.Timelapse) []string {
	return []string{
		//"-r", "60",  // using might be causing the "Past duration 0.999992 too large" errors?
		"-framerate", "60",
		"-start_number", fmt.Sprintf("%d", t.Start),
		"-i", t.GetFFmpegInputPath(),
		"-c:v", "libx264",
		"-preset", "slow",
		"-crf", "17",
		"-s", "1920x1080",
		"-progress", "/dev/stdout",
		t.GetOutputFullPath("1080p-test.mp4"),
	}
}

func (f *configFake) GetDebugFullPath(t *filebrowse.Timelapse) string {
	return t.GetOutputFullPath("1080p-test.mp4.log")
}

func (f *configFake) GetDebugPath(t *filebrowse.Timelapse) string {
	return t.GetOutputPath("1080p-test.mp4.log")
}

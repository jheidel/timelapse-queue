package engine

import (
	"fmt"

	"timelapse-queue/filebrowse"
)

type Bounds struct {
	X, Y, Width, Height int
}

func (b *Bounds) GetCropArg() string {
	return fmt.Sprintf("crop=%d:%d:%d:%d", b.Width, b.Height, b.X, b.Y)
}

type Config interface {
	GetArgs(timelapse *filebrowse.Timelapse) []string
	Validate(timelapse *filebrowse.Timelapse) error
	GetDebugFullPath(timelapse *filebrowse.Timelapse) string
	GetDebugPath(timelapse *filebrowse.Timelapse) string
}

type configFake struct {
	Bounds *Bounds
}

// TODO: rest of this class. Some things like progress implementation are
// ffmpeg internal and should be pushed into ffmpeg.go.

func (f *configFake) GetArgs(t *filebrowse.Timelapse) []string {
	return []string{
		//"-r", "60",  // using might be causing the "Past duration 0.999992 too large" errors?
		"-framerate", "60",
		"-start_number", fmt.Sprintf("%d", t.Start),
		"-i", t.GetFFmpegInputPath(),
		"-vf", f.Bounds.GetCropArg(),
		"-c:v", "libx264",
		"-preset", "slow",
		"-crf", "17",
		"-s", "1920x1080",
		"-progress", "/dev/stdout",
		t.GetOutputFullPath("1080p-test.mp4"),
	}
}

func (f *configFake) Validate(t *filebrowse.Timelapse) error {
	if f.Bounds.Width < 1920 || f.Bounds.Height < 1080 {
		return fmt.Errorf("Image must be at least 1920x1080")
	}

	// TODO make sure X and Y are in bounds of timelapse.
	return nil
}

func (f *configFake) GetDebugFullPath(t *filebrowse.Timelapse) string {
	return t.GetOutputFullPath("1080p-test.mp4.log")
}

func (f *configFake) GetDebugPath(t *filebrowse.Timelapse) string {
	return t.GetOutputPath("1080p-test.mp4.log")
}

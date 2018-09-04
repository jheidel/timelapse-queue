package engine

import (
	"fmt"
	"image"

	"timelapse-queue/filebrowse"
)

type Config interface {
	Validate(timelapse *filebrowse.Timelapse) error

	GetDebugFullPath(timelapse *filebrowse.Timelapse) string
	GetDebugPath(timelapse *filebrowse.Timelapse) string

	GetRegion() image.Rectangle

	//Images(timelapse *filebrowse.Timelapse) <-chan *image.RGBA
}

type configFake struct {
	Region image.Rectangle
}

func (f *configFake) GetRegion() image.Rectangle {
	return f.Region
}

// TODO: rest of this class. Some things like progress implementation are
// ffmpeg internal and should be pushed into ffmpeg.go.

func (f *configFake) Validate(t *filebrowse.Timelapse) error {
	width := f.Region.Max.X - f.Region.Min.X
	height := f.Region.Max.Y - f.Region.Min.Y
	if width < 1920 || height < 1080 {
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

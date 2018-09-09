package engine

import (
	"context"
	"fmt"
	"image"
	"time"

	log "github.com/sirupsen/logrus"
	"timelapse-queue/filebrowse"
)

type Config interface {
	// The basename for the output timelapse.
	GetFilename() string
	// The basename for the output debug file.
	GetDebugFilename() string
	// The desired cropping region.
	GetRegion() image.Rectangle
	// Gets the start & end of the sequence.
	GetStartEnd() (int, int)

	// Gets the expected number of output frames in the sequence (to compute progress)
	GetExpectedFrames() int
}

type baseConfig struct {
	Path                 string
	X, Y, Width, Height  int
	OutputName           string
	StartFrame, EndFrame int
}

func (f *baseConfig) GetRegion() image.Rectangle {
	return image.Rectangle{
		Min: image.Point{
			X: f.X,
			Y: f.Y,
		},
		Max: image.Point{
			X: f.X + f.Width,
			Y: f.Y + f.Height,
		},
	}
}

func getSampleImageBounds(pctx context.Context, t *filebrowse.Timelapse, start int) (image.Rectangle, error) {
	ctx, cancelf := context.WithTimeout(pctx, 10*time.Second)
	defer cancelf()
	imagec, errc := t.Images(ctx, start, 0)
	select {
	case img := <-imagec:
		return img.Rect, nil
	case err := <-errc:
		return image.Rectangle{}, err
	}
}

func (f *baseConfig) Validate(ctx context.Context, t *filebrowse.Timelapse) error {
	if f.OutputName == "" {
		return fmt.Errorf("missing output filename")
	}

	if f.StartFrame < 0 || f.StartFrame >= t.Count {
		return fmt.Errorf("start frame out of bounds")
	}
	if f.EndFrame < 0 || f.EndFrame >= t.Count {
		return fmt.Errorf("end frame out of bounds")
	}
	if f.StartFrame >= f.EndFrame {
		return fmt.Errorf("start frame must come before end frame")
	}

	r := f.GetRegion()
	if r.Dx() < 1920 || r.Dy() < 1080 {
		return fmt.Errorf("Image must be at least 1920x1080")
	}

	ir, err := getSampleImageBounds(ctx, t, f.StartFrame)
	if err != nil {
		return err
	}

	log.Infof("Region is %v image is %v", r, ir)

	if !(r.Min.X >= ir.Min.X && r.Min.Y >= ir.Min.Y &&
		r.Min.X <= ir.Max.X && r.Min.Y <= ir.Max.Y &&
		r.Max.X >= ir.Min.X && r.Max.Y >= ir.Min.Y &&
		r.Max.X <= ir.Max.X && r.Max.Y <= ir.Max.Y) {
		return fmt.Errorf("crop rectangle out of bounds of source image")
	}
	return nil
}

func (f *baseConfig) GetFilename() string {
	return f.OutputName + ".mp4"
}

func (f *baseConfig) GetDebugFilename() string {
	return f.GetFilename() + ".log"
}

func (f *baseConfig) GetStartEnd() (int, int) {
	return f.StartFrame, f.EndFrame
}

func (f *baseConfig) GetExpectedFrames() int {
	return f.EndFrame + 1 - f.StartFrame
}

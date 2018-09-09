package engine

import (
	"context"
	"fmt"
	"image"
	"time"

	"timelapse-queue/filebrowse"
	"timelapse-queue/process"
)

// TODO not a huge fan of this interface being here...
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

	// Get convert options
	GetConvertOptions() *ConvertOptions
}

type baseConfig struct {
	Path                   string
	X, Y, Width, Height    int
	OutputName             string
	StartFrame, EndFrame   int
	ProfileCPU, ProfileMem bool

	Stack          bool
	StackWindow    int
	StackSkipCount int
	StackMode      string
}

func (f *baseConfig) GetConvertOptions() *ConvertOptions {
	return &ConvertOptions{
		Stack:          f.Stack,
		StackWindow:    f.StackWindow,
		StackSkipCount: f.StackSkipCount,
		StackMode:      f.StackMode,
		ProfileCPU:     f.ProfileCPU,
		ProfileMem:     f.ProfileMem,
	}
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
		return fmt.Errorf("selected region must be at least 1920x1080")
	}
	ir, err := getSampleImageBounds(ctx, t, f.StartFrame)
	if err != nil {
		return fmt.Errorf("failed to load sample frame: %v", err)
	}

	if !(r.Min.X >= ir.Min.X && r.Min.Y >= ir.Min.Y &&
		r.Min.X <= ir.Max.X && r.Min.Y <= ir.Max.Y &&
		r.Max.X >= ir.Min.X && r.Max.Y >= ir.Min.Y &&
		r.Max.X <= ir.Max.X && r.Max.Y <= ir.Max.Y) {
		return fmt.Errorf("crop rectangle out of bounds of source image")
	}

	if f.ProfileCPU && f.ProfileMem {
		return fmt.Errorf("only one profile mode at a time is supported")
	}

	if f.Stack {
		smax := (f.EndFrame - f.StartFrame + 1)
		if f.StackWindow < 0 || f.StackWindow > smax {
			return fmt.Errorf("stacking window out of range 0..%d", smax)
		}
		if f.StackSkipCount < 0 || f.StackSkipCount > smax || f.StackSkipCount > f.StackWindow {
			return fmt.Errorf("stacking skip count out of range")
		}
		if process.GetMergerByName(f.StackMode) == nil {
			return fmt.Errorf("invalid stack mode %v", f.StackMode)
		}
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
	frames := f.EndFrame + 1 - f.StartFrame
	if f.Stack {
		// Stacking will add additonal frames to the output.
		frames += f.StackWindow
	}
	return frames
}

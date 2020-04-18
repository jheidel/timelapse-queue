package process

import (
	"context"
	"image"
	"math"

	"github.com/BurntSushi/graphics-go/graphics"
)

type Rotate struct {
	Degrees int
}

func toRadians(deg int) float64 {
	return float64(deg) * math.Pi / 180
}

func SizeAfterRotate(sz image.Rectangle, degrees int) image.Rectangle {
	if degrees < 0 {
		degrees *= -1
	}
	f := degrees > 90
	if f {
		degrees -= 90
	}
	rad := toRadians(degrees)
	w := int(float64(sz.Dx())*math.Cos(rad) + float64(sz.Dy())*math.Sin(rad))
	h := int(float64(sz.Dx())*math.Sin(rad) + float64(sz.Dy())*math.Cos(rad))
	if f {
		t := w
		w = h
		h = t
	}
	return image.Rect(0, 0, w, h)
}

func (r *Rotate) rotate(in *image.RGBA) *image.RGBA {
	dst := image.NewRGBA(SizeAfterRotate(in.Bounds(), r.Degrees))
	graphics.Rotate(dst, in, &graphics.RotateOptions{toRadians(r.Degrees)})
	return dst
}

func (r *Rotate) Process(ctx context.Context, inc <-chan *image.RGBA, errc chan error) (<-chan *image.RGBA, chan error) {
	outc := make(chan *image.RGBA)
	go func() {
		defer close(outc)
		for img := range inc {
			out := r.rotate(img)
			select {
			case <-ctx.Done():
				return
			case outc <- out:
			}
		}
	}()
	return outc, errc
}

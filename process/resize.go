package process

import (
	"context"
	"fmt"
	"image"

	"github.com/nfnt/resize"
)

type Resizer struct {
	Size image.Point
}

func (r *Resizer) resize(in *image.RGBA) (*image.RGBA, error) {
	// TODO check performance?
	small := resize.Resize(uint(r.Size.X), uint(r.Size.Y), in, resize.Bicubic)
	//small := resize.Resize(uint(r.Size.X), uint(r.Size.Y), in, resize.NearestNeighbor)
	img, ok := small.(*image.RGBA)
	if !ok {
		return nil, fmt.Errorf("Inexpected image format after resize")
	}
	return img, nil
}

func (r *Resizer) Process(ctx context.Context, inc <-chan *image.RGBA, errc chan error) (<-chan *image.RGBA, chan error) {
	outc := make(chan *image.RGBA)
	go func() {
		defer close(outc)
		for in := range inc {
			out, err := r.resize(in)
			if err != nil {
				errc <- err
				return
			}
			select {
			case <-ctx.Done():
				return
			case outc <- out:
			}
		}
	}()
	return outc, errc
}

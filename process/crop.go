package process

import (
	"context"
	"fmt"
	"image"
)

type Crop struct {
	Region image.Rectangle
}

func (c *Crop) crop(in *image.RGBA) (*image.RGBA, error) {
	img := in.SubImage(c.Region)
	out, ok := img.(*image.RGBA)
	if !ok {
		return nil, fmt.Errorf("Output of subimage not RGBA")
	}
	return out, nil
}

func (c *Crop) Process(ctx context.Context, inc <-chan *image.RGBA, errc chan error) (<-chan *image.RGBA, chan error) {
	outc := make(chan *image.RGBA)
	go func() {
		defer close(outc)
		for img := range inc {
			out, err := c.crop(img)
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

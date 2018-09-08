package process

import (
	"image"
)

type Stacker struct {
	Overlap int
	Merger  Merger
}

type Merger interface {
	Blend(img1, img2 *image.RGBA) *image.RGBA
}

type Lighten struct{}

func (l *Lighten) Blend(img1, img2 *image.RGBA) *image.RGBA {
	m := image.NewRGBA(img1.Rect)
	sz := len(img1.Pix)
	for i := 0; i < sz; i++ {
		// Maximum pixel value
		if img1.Pix[i] > img2.Pix[i] {
			m.Pix[i] = img1.Pix[i]
		} else {
			m.Pix[i] = img2.Pix[i]
		}
	}
	return m
}

// BEFORE: 19 minutes for merge of 60
func (s *Stacker) stack(imgs []*image.RGBA) *image.RGBA {
	m := imgs[0]
	for _, v := range imgs[1:] {
		m = s.Merger.Blend(m, v)
	}
	return m
}

func (s *Stacker) Process(inc <-chan *image.RGBA, errc chan error) (<-chan *image.RGBA, chan error) {
	outc := make(chan *image.RGBA)
	go func() {
		defer close(outc)

		// TODO support buffer
		hist := []*image.RGBA{}

		for img := range inc {
			hist = append(hist, img)
			if len(hist) > s.Overlap {
				hist = hist[1:]
			}
			outc <- s.stack(hist)
		}
	}()
	return outc, errc
}

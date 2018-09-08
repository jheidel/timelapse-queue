package process

import (
	"image"

	log "github.com/sirupsen/logrus"
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

func (s *Stacker) Process(inc <-chan *image.RGBA, errc chan error) (<-chan *image.RGBA, chan error) {
	outc := make(chan *image.RGBA)
	go func() {
		defer close(outc)

		buf := &Buffer{}
		frame := 0
		window := []int{}

		for img := range inc {
			buf.Add(frame, img)

			window = append(window, frame)
			if len(window) > s.Overlap {
				buf.RemoveOld(window[0])
				window = window[1:]
			}

			outc <- buf.Generate(window, s.Merger)
			frame += 1
		}

		for len(window) > 1 {
			buf.RemoveOld(window[0])
			window = window[1:]
			outc <- buf.Generate(window, s.Merger)
		}

		log.Infof("stacker exit")
	}()
	return outc, errc
}

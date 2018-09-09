package process

import (
	"image"

	log "github.com/sirupsen/logrus"
)

type Stacker struct {
	// Number of frames to overlap; 0 overlaps all frames.
	Overlap int
	// Number of frames to skip during overlap; 0 disables.
	Skip int

	// The underlying merge processor.
	Merger Merger
}

type Merger interface {
	Blend(img1, img2 *image.RGBA) *image.RGBA
}

func GetMergerByName(mode string) Merger {
	if mode == "lighten" {
		return &Lighten{}
	}
	if mode == "darken" {
		return &Darken{}
	}
	return nil
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

type Darken struct{}

func (l *Darken) Blend(img1, img2 *image.RGBA) *image.RGBA {
	m := image.NewRGBA(img1.Rect)
	sz := len(img1.Pix)
	for i := 0; i < sz; i++ {
		// Minimum pixel value
		if img1.Pix[i] < img2.Pix[i] {
			m.Pix[i] = img1.Pix[i]
		} else {
			m.Pix[i] = img2.Pix[i]
		}
	}
	return m
}

func (s *Stacker) applySkipToWindow(in []int) []int {
	if s.Skip == 0 {
		// Skipping disabled.
		return in
	}

	out := []int{}
	for _, v := range in {
		if v%s.Skip == 0 {
			out = append(out, v)
		}
	}
	if out[len(out)-1] != in[len(in)-1] {
		out = append(out, in[len(in)-1])
	}
	return out
}

func (s *Stacker) overlapWindow(inc <-chan *image.RGBA, outc chan<- *image.RGBA) {
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

		outc <- buf.Generate(s.applySkipToWindow(window), s.Merger)
		frame += 1
	}

	for len(window) > 1 {
		buf.RemoveOld(window[0])
		window = window[1:]
		outc <- buf.Generate(window, s.Merger)
	}
}

func (s *Stacker) overlapAll(inc <-chan *image.RGBA, outc chan<- *image.RGBA) {
	var hist *image.RGBA
	for img := range inc {
		if hist == nil {
			hist = img
		} else {
			hist = s.Merger.Blend(img, hist)
		}
		outc <- hist
	}
}

func (s *Stacker) Process(inc <-chan *image.RGBA, errc chan error) (<-chan *image.RGBA, chan error) {
	outc := make(chan *image.RGBA)
	go func() {
		defer close(outc)

		if s.Overlap > 0 {
			s.overlapWindow(inc, outc)
		} else {
			s.overlapAll(inc, outc)
		}
		log.Infof("stacker done")
	}()
	return outc, errc
}

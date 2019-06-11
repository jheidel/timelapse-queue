package filebrowse

import (
	"context"
	"fmt"
	"image"
	"os"
	"path/filepath"

	"github.com/pixiv/go-libjpeg/jpeg"
)

type Timelapse struct {
	// The base name of the first image in the sequence.
	Name string
	// The path to the first image in the sequence, relative to the file browser
	// root. Useful to ID the timelapse.
	Path string
	// ParentPath is the path to the directory, relative to the file browser root.
	ParentPath string
	// The text component preceding the numbers.
	Prefix string
	// The file extension component.
	Ext string
	// NumLen is the string length of the number component.
	NumLen int
	// Count is the number of images in the sequence.
	Count int
	// Start is the index of the first timelapse.
	Start int

	// DurationString is the length of the timelapse as a human readable string.
	DurationString string

	browser *FileBrowser
}

// GetOutputFullPath returns a path that can be used for file output with the given basename.
func (t *Timelapse) GetOutputFullPath(base string) string {
	return filepath.Join(t.browser.Root, t.GetOutputPath(base))
}

// GetOutputPath returns a relative path that can be used for file output with the given basename.
func (t *Timelapse) GetOutputPath(base string) string {
	parent, _ := filepath.Split(t.Path)
	return filepath.Join(parent, base)
}

func (t *Timelapse) GetPathForIndex(idx int) string {
	basef := fmt.Sprintf("%s%%%02dd.%s", t.Prefix, t.NumLen, t.Ext)
	base := fmt.Sprintf(basef, t.Start+idx)
	return t.GetOutputFullPath(base)
}

func (t *Timelapse) getImage(num int) (*image.RGBA, error) {
	f, err := os.Open(t.GetPathForIndex(num))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Requires libjpeg-turbo
	img, err := jpeg.DecodeIntoRGBA(f, &jpeg.DecoderOptions{})
	if err != nil {
		return nil, err
	}
	return img, nil
}

// Images produces a stream of images for this timelapse.
// Optionally supply non-zero start & end for bounded timelapse.
func (t *Timelapse) Images(ctx context.Context, start, end int) (<-chan *image.RGBA, chan error) {
	errc := make(chan error, 1)
	imagec := make(chan *image.RGBA)
	go func() {
		defer close(imagec)
		defer close(errc)
		if end == 0 {
			end = t.Count - 1
		}
		for i := start; i <= end; i++ {
			img, err := t.getImage(i)
			if err != nil {
				errc <- err
				return
			}
			if ctx.Err() != nil {
				return
			}
			imagec <- img
		}
	}()
	return imagec, errc
}

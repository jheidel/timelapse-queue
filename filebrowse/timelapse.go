package filebrowse

import (
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

// GetFFmpegInputPath returns a full path that can be used for FFmpegInput.
func (t *Timelapse) GetFFmpegInputPath() string {
	base := fmt.Sprintf("%s%%%02dd.%s", t.Prefix, t.NumLen, t.Ext)
	return t.GetOutputFullPath(base)
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
func (t *Timelapse) Images() (<-chan *image.RGBA, chan error) {
	errc := make(chan error, 1)
	imagec := make(chan *image.RGBA)
	go func() {
		defer close(imagec)
		defer close(errc)
		for i := 0; i < t.Count; i++ {
			img, err := t.getImage(i)
			if err != nil {
				errc <- err
				return
			}
			imagec <- img
		}
	}()
	return imagec, errc
}

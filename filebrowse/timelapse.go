package filebrowse

import (
	"context"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"time"

	"github.com/pixiv/go-libjpeg/jpeg"
)

// ITimelapse is a generic timelapse interface.
type ITimelapse interface {
	// TimelapseName is the human-readable name of this timelapse.
	TimelapseName() string
	// Path can be used for /image
	ImagePath() string
	GetPathForIndex(idx int) string
	GetOutputFullPath(base string) string
	ImageCount() int
	View() *TimelapseView
}

type TimelapseView struct {
	OutputPath     string
	Count          int
	DurationString string
}

// Timelapse represents a sequence of images on disk.
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

	browser *FileBrowser
}

func toDuration(t ITimelapse) string {
	fps := 60
	dur := time.Second * time.Duration(t.ImageCount()) / time.Duration(fps)
	return dur.Truncate(100 * time.Millisecond).String()
}

func (t *Timelapse) TimelapseName() string {
	return t.Name
}

func (t *Timelapse) ImagePath() string {
	return t.Path
}

func (t *Timelapse) View() *TimelapseView {
	return &TimelapseView{
		OutputPath:     t.ParentPath,
		Count:          t.ImageCount(),
		DurationString: toDuration(t),
	}
}

// GetOutputFullPath returns a path that can be used for file output with the given basename.
func (t *Timelapse) GetOutputFullPath(base string) string {
	parent, _ := filepath.Split(t.Path)
	rel := filepath.Join(parent, base)
	return filepath.Join(t.browser.Root, rel)
}

func (t *Timelapse) ImageCount() int {
	return t.Count
}

func (t *Timelapse) GetPathForIndex(idx int) string {
	if idx < 0 || idx >= t.Count {
		panic("out of bounds")
	}
	basef := fmt.Sprintf("%s%%%02dd.%s", t.Prefix, t.NumLen, t.Ext)
	base := fmt.Sprintf(basef, t.Start+idx)
	return t.GetOutputFullPath(base)
}

func getImage(path string) (*image.RGBA, error) {
	f, err := os.Open(path)
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

// ImagePaths produces a stream of paths for this timelapse.
// Paths are absolute.
// Optionally supply non-zero start & end for bounded timelapse.
func ImagePaths(t ITimelapse, start, end, skip int) <-chan string {
	pathc := make(chan string)
	go func() {
		defer close(pathc)
		if end == 0 {
			end = t.ImageCount() - 1
		}
		for i := start; i <= end; i += skip {
			pathc <- t.GetPathForIndex(i)
		}
	}()
	return pathc
}

// Images produces a stream of images for this timelapse.
// Optionally supply non-zero start & end for bounded timelapse.
func Images(ctx context.Context, t ITimelapse, start, end, skip int) (<-chan *image.RGBA, chan error) {
	errc := make(chan error, 1)
	imagec := make(chan *image.RGBA)
	go func() {
		defer close(imagec)
		defer close(errc)
		for path := range ImagePaths(t, start, end, skip) {
			img, err := getImage(path)
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

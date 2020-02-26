package filebrowse

import (
	"fmt"
	"image"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/nfnt/resize"
	"github.com/pixiv/go-libjpeg/jpeg"
)

const (
	ThumbSize = 200
)

type ImageHost struct {
	Browser *FileBrowser
}

func (h *ImageHost) writeImage(rel string, idx string, thumb bool, w http.ResponseWriter) error {
	var err error
	i := 0
	if idx != "" {
		i, err = strconv.Atoi(idx)
		if err != nil {
			return err
		}
	}

	t, err := h.Browser.GetTimelapse(rel)
	if err != nil {
		return err
	}

	count := t.ImageCount()
	if i < 0 || i >= count {
		return fmt.Errorf("index out of timelapse range %d to %d", 0, count-1)
	}

	imf, err := os.Open(t.GetPathForIndex(i))
	if err != nil {
		return err
	}

	if !thumb {
		_, err := io.Copy(w, imf)
		if err != nil {
			return err
		}
		return nil
	}

	// Downsample prior to resize.
	r := image.Rectangle{
		Min: image.Point{X: 0, Y: 0},
		Max: image.Point{X: ThumbSize, Y: ThumbSize},
	}
	im, err := jpeg.Decode(imf, &jpeg.DecoderOptions{ScaleTarget: r})
	if err != nil {
		return err
	}

	im = resize.Thumbnail(ThumbSize, ThumbSize, im, resize.Bilinear)

	w.Header().Set("Content-Type", "image/jpeg")
	err = jpeg.Encode(w, im, &jpeg.EncoderOptions{Quality: 90})
	if err != nil {
		return err
	}

	return nil
}

func (h *ImageHost) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	rel := r.Form.Get("path")
	idx := r.Form.Get("index")
	thumb := r.Form.Get("thumb") != ""
	w.Header().Set("Content-Type", "image/jpeg")
	if err := h.writeImage(rel, idx, thumb, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

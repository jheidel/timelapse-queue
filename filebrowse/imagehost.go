package filebrowse

import (
	"image"
	"image/jpeg"
	_ "image/png"
	"net/http"
	"os"

	"github.com/nfnt/resize"
)

const (
	ThumbSize = 200
)

type ImageHost struct {
	Browser *FileBrowser
}

func (h *ImageHost) writeImage(rel string, thumb bool, w http.ResponseWriter) error {
	path, err := h.Browser.GetFullPath(rel)
	if err != nil {
		return err
	}

	imf, err := os.Open(path)
	if err != nil {
		return err
	}

	im, _, err := image.Decode(imf)
	if err != nil {
		return err
	}

	if thumb {
		im = resize.Thumbnail(ThumbSize, ThumbSize, im, resize.Bilinear)
	}

	w.Header().Set("Content-Type", "image/jpeg")
	err = jpeg.Encode(w, im, nil)
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
	thumb := r.Form.Get("thumb") != ""
	if err := h.writeImage(rel, thumb, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

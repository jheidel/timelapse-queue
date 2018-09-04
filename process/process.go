package process

import (
	"image"
)

type Process interface {
	Process(<-chan *image.RGBA, chan error) (<-chan *image.RGBA, chan error)
}

package filebrowse

import (
	"fmt"
	"path/filepath"
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

// GetOuputFullPath returns a path that can be used for file output with the given basename.
func (t *Timelapse) GetOutputFullPath(base string) string {
	return filepath.Join(t.browser.Root, t.GetOutputPath(base))
}

// GetOuputPath returns a relative path that can be used for file output with the given basename.
func (t *Timelapse) GetOutputPath(base string) string {
	parent, _ := filepath.Split(t.Path)
	return filepath.Join(parent, base)
}

// GetFFmpegInputPath returns a full path that can be used for FFmpegInput.
func (t *Timelapse) GetFFmpegInputPath() string {
	base := fmt.Sprintf("%s%%%02dd.%s", t.Prefix, t.NumLen, t.Ext)
	return t.GetOutputFullPath(base)
}

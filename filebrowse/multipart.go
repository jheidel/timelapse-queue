package filebrowse

import (
	"fmt"
)

type MultipartTimelapse struct {
	Parts []*Timelapse
}

func (t *MultipartTimelapse) GetPathForIndex(idx int) string {
	for _, p := range t.Parts {
		if idx < p.Count {
			return p.GetPathForIndex(idx)
		}
		idx -= p.Count
	}
	panic("out of bounds for multipart")
}

func (t *MultipartTimelapse) GetOutputFullPath(base string) string {
	return t.Parts[0].GetOutputFullPath(base)
}

func (t *MultipartTimelapse) ImageCount() int {
	s := 0
	for _, p := range t.Parts {
		s += p.Count
	}
	return s
}

func (t *MultipartTimelapse) View() *TimelapseView {
	return &TimelapseView{
		OutputPath:     t.Parts[0].ParentPath,
		Count:          t.ImageCount(),
		DurationString: toDuration(t),
	}
}

func (t *MultipartTimelapse) TimelapseName() string {
	return fmt.Sprintf("%s (%d parts)", t.Parts[0].Name, len(t.Parts))
}

func (t *MultipartTimelapse) ImagePath() string {
	return t.Parts[0].Path
}

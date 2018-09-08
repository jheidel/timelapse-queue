package process

import (
	"image"
	"sort"
)

// Caches past image merges in order to significantly speed up overlap merging.
// Assumes all image merge operations are commutative.

type BufferItem struct {
	// The input frames that are combined to make this frame, ordered.
	Frames []int
	Result *image.RGBA
}

type Buffer struct {
	items []*BufferItem
}

func (b *Buffer) Add(frame int, img *image.RGBA) {
	b.items = append(b.items, &BufferItem{
		Frames: []int{frame},
		Result: img,
	})
}

// Remove drops all buffered items at or before the given frame.
func (b *Buffer) RemoveOld(frame int) {
	items := []*BufferItem{}
	for _, item := range b.items {
		if item.Frames[0] <= frame {
			continue
		}
		items = append(items, item)
	}
	b.items = items
}

// Whether a is a subset of b. Both lists must be sorted.
// https://stackoverflow.com/questions/18879109/subset-check-with-slices-in-go
func subset(a, b []int) bool {
	for len(a) > 0 {
		switch {
		case len(b) == 0:
			return false
		case a[0] == b[0]:
			a = a[1:]
			b = b[1:]
		case a[0] < b[0]:
			return false
		case a[0] > b[0]:
			b = b[1:]
		}
	}
	return true
}

// Removes elements a from b. Assumes both lists are sorted.
func remove(a, b []int) []int {
	result := []int{}
	for _, v := range b {
		if len(a) > 0 && v == a[0] {
			a = a[1:]
		} else {
			result = append(result, v)
		}
	}
	return result
}

// Gets the buffer item which covers the largest subset of the needed frames.
func (b *Buffer) getLargestSubset(frames []int) *BufferItem {
	possible := []*BufferItem{}
	for _, item := range b.items {
		if subset(item.Frames, frames) {
			possible = append(possible, item)
		}
	}

	max := 0
	var largest *BufferItem
	for _, p := range possible {
		if len(p.Frames) > max {
			max = len(p.Frames)
			largest = p
		}
	}

	if largest == nil {
		panic("no matching item")
	}
	return largest
}

// Gets the minimal set of buffered historical frames to build this frame range.
func (b *Buffer) getSpanning(frames []int) []*BufferItem {
	needed := frames
	result := []*BufferItem{}
	for len(needed) > 0 {
		item := b.getLargestSubset(needed)
		result = append(result, item)
		needed = remove(item.Frames, needed)
	}

	// Ensure components are in ascending frame order.
	sort.Slice(result, func(i, j int) bool {
		return result[i].Frames[0] < result[j].Frames[0]
	})
	return result
}

func (b *Buffer) Generate(frames []int, merger Merger) *image.RGBA {
	spans := b.getSpanning(frames)

	var item *BufferItem

	// Iterate backwards, joining together and adding to cache.
	for len(spans) > 0 {
		tail := spans[len(spans)-1]
		spans = spans[:len(spans)-1]

		if item == nil {
			item = tail
		} else {
			new := &BufferItem{
				Frames: append(tail.Frames, item.Frames...),
				Result: merger.Blend(item.Result, tail.Result),
			}
			b.items = append(b.items, new)
			item = new
		}
	}

	if item == nil {
		panic("could not build frame")
	}
	return item.Result
}

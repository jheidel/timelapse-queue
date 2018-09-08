package process

import (
	"image"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestSpanning(t *testing.T) {
	tests := []struct {
		name   string
		cached [][]int
		input  []int
		want   [][]int
	}{
		{
			name:   "individual frame",
			cached: [][]int{{1}},
			input:  []int{1},
			want:   [][]int{{1}},
		},
		{
			name:   "multiple indivdual frames",
			cached: [][]int{{1}, {2}, {3}},
			input:  []int{1, 2, 3},
			want:   [][]int{{1}, {2}, {3}},
		},
		{
			name:   "larger",
			cached: [][]int{{1}, {2, 3}, {3}},
			input:  []int{1, 2, 3},
			want:   [][]int{{1}, {2, 3}},
		},
		{
			name:   "even larger",
			cached: [][]int{{1}, {2}, {3}, {4}, {5}, {6}, {3, 4}, {2, 3, 4}, {5, 6}},
			input:  []int{1, 2, 3, 4, 5, 6},
			want:   [][]int{{1}, {2, 3, 4}, {5, 6}},
		},
		{
			name:   "desired skip",
			cached: [][]int{{1}, {2}, {3}, {4}, {2, 3}, {2, 3, 4}, {3, 4}},
			input:  []int{2, 4},
			want:   [][]int{{2}, {4}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buffer := &Buffer{}
			for _, frames := range test.cached {
				buffer.items = append(buffer.items, &BufferItem{
					Frames: frames,
				})
			}
			s := buffer.getSpanning(test.input)
			got := [][]int{}
			for _, v := range s {
				got = append(got, v.Frames)
			}
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("spanning diffs: %v", diff)
			}
		})
	}
}

type fakeMerger struct {
}

func (f *fakeMerger) Blend(img1, img2 *image.RGBA) *image.RGBA {
	return nil
}

func TestGenerate(t *testing.T) {
	tests := []struct {
		name   string
		before [][]int
		input  []int
		after  [][]int
	}{
		{
			name:   "individual frame",
			before: [][]int{{1}},
			input:  []int{1},
			after:  [][]int{{1}},
		},
		{
			name:   "frame 4",
			before: [][]int{{1}, {2}, {3}, {4}},
			input:  []int{1, 2, 3, 4},
			after:  [][]int{{1}, {2}, {3}, {4}, {3, 4}, {2, 3, 4}, {1, 2, 3, 4}},
		},
		{
			name:   "frame 5",
			before: [][]int{{2}, {3}, {4}, {3, 4}, {2, 3, 4}, {5}},
			input:  []int{2, 3, 4, 5},
			after:  [][]int{{2}, {3}, {4}, {3, 4}, {2, 3, 4}, {5}, {2, 3, 4, 5}},
		},
		{
			name:   "frame 6",
			before: [][]int{{3}, {4}, {3, 4}, {5}, {6}},
			input:  []int{3, 4, 5, 6},
			after:  [][]int{{3}, {4}, {3, 4}, {5}, {6}, {5, 6}, {3, 4, 5, 6}},
		},
		{
			name:   "frame 7",
			before: [][]int{{4}, {5}, {6}, {5, 6}, {7}},
			input:  []int{4, 5, 6, 7},
			after:  [][]int{{4}, {5}, {6}, {5, 6}, {7}, {5, 6, 7}, {4, 5, 6, 7}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buffer := &Buffer{}
			for _, frames := range test.before {
				buffer.items = append(buffer.items, &BufferItem{
					Frames: frames,
				})
			}
			buffer.Generate(test.input, &fakeMerger{})
			got := [][]int{}
			for _, v := range buffer.items {
				got = append(got, v.Frames)
			}
			if diff := cmp.Diff(test.after, got); diff != "" {
				t.Errorf("buffer diffs: %v", diff)
			}
		})
	}
}

func TestAddRemove(t *testing.T) {
	tests := []struct {
		name   string
		before [][]int
		remove int
		add    int
		after  [][]int
	}{
		{
			name:   "basic",
			before: [][]int{{3}, {4}, {3, 4}, {5}, {6}, {5, 6}, {3, 4, 5, 6}},
			remove: 3,
			add:    7,
			after:  [][]int{{4}, {5}, {6}, {5, 6}, {7}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buffer := &Buffer{}
			for _, frames := range test.before {
				buffer.items = append(buffer.items, &BufferItem{
					Frames: frames,
				})
			}
			buffer.RemoveOld(test.remove)
			buffer.Add(test.add, nil)
			got := [][]int{}
			for _, v := range buffer.items {
				got = append(got, v.Frames)
			}
			if diff := cmp.Diff(test.after, got); diff != "" {
				t.Errorf("buffer diffs: %v", diff)
			}
		})
	}
}

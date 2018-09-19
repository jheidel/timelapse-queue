package engine

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Profile struct {
	Name       string
	Width      int
	Height     int
	FFmpegArgs []string
}

// Profiles defines the possible output configurations.
var Profiles = []*Profile{
	{
		Name:   "1080p (1920x1080)",
		Width:  1920,
		Height: 1080,
		FFmpegArgs: []string{
			"-level:v", "4.2",
			"-profile:v", "high",
			"-pix_fmt", "yuv420p",
		},
	},
	{
		Name:   "4k (3840x2160)",
		Width:  3840,
		Height: 2160,
		FFmpegArgs: []string{
			"-level:v", "5.2",
			"-profile:v", "high",
			"-pix_fmt", "yuv420p",
		},
	},
	{
		Name:   "12MP (4000x3000)",
		Width:  4000,
		Height: 3000,
		FFmpegArgs: []string{
			"-level:v", "6.2",
			"-profile:v", "high",
			"-pix_fmt", "yuv420p",
		},
	},
}

func GetProfileByName(name string) (*Profile, error) {
	if name == "" {
		return nil, fmt.Errorf("Output resolution profile not specified")
	}
	for _, p := range Profiles {
		if p.Name == name {
			return p, nil
		}
	}
	return nil, fmt.Errorf("Unknown output resolution profile %q", name)
}

func ServeProfiles(w http.ResponseWriter, r *http.Request) {
	js, err := json.Marshal(Profiles)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

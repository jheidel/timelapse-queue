package filebrowse

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	cache "github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
)

const (
	// FilesystemCacheDuration controls the maximum allowed staleness of filesystem entries.
	FilesystemCacheDuration = 15 * time.Second
)

var (
	timelapseRE = regexp.MustCompile(`^([^\d]*)(\d+)\.(\w+)$`)

	allowedEXT = []string{"jpg", "png"}
)

type FileBrowser struct {
	// Root is the base of the file system to serve up
	Root string

	// listCache is the cache used for (possibly expensive) file list operations
	listCache *cache.Cache
}

func NewFileBrowser(root string) *FileBrowser {
	c := cache.New(FilesystemCacheDuration, 5*time.Minute)
	return &FileBrowser{
		Root:      root,
		listCache: c,
	}
}

type Directory struct {
	Name string
	Path string
}
type Response struct {
	Parents    []*Directory
	Dirs       []*Directory
	Timelapses []*Timelapse
}

func (f *FileBrowser) GetFullPath(p string) (string, error) {
	root, err := filepath.EvalSymlinks(f.Root)
	if err != nil {
		return "", err
	}
	b, err := filepath.EvalSymlinks(filepath.Join(root, p))
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(b, root) {
		return "", errors.New("permission denied, not in root")
	}
	return b, nil
}

func (f *FileBrowser) GetTimelapse(p string) (*Timelapse, error) {
	dir, name := path.Split(p)

	contents, err := f.listPath(dir)
	if err != nil {
		return nil, err
	}

	for _, t := range contents.Timelapses {
		if t.Name == name {
			return t, nil
		}
	}
	return nil, fmt.Errorf("timelapse %v not found in %v", name, dir)
}

func (f *FileBrowser) listPath(p string) (*Response, error) {
	root, err := filepath.EvalSymlinks(f.Root)
	if err != nil {
		return nil, err
	}
	b, err := filepath.EvalSymlinks(filepath.Join(root, p))
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(b, root) {
		return nil, errors.New("permission denied, not in root")
	}

	var files []os.FileInfo
	if v, found := f.listCache.Get(b); found {
		files = v.([]os.FileInfo)
	} else {
		start := time.Now()
		v, err := ioutil.ReadDir(b)
		if err != nil {
			return nil, err
		}
		elapsed := time.Now().Sub(start).Truncate(time.Millisecond)
		log.Infof("Read of %v returned %d entries in %v", b, len(v), elapsed)
		f.listCache.Set(b, v, cache.DefaultExpiration)
		files = v
	}

	r := &Response{}

	tmap := make(map[string]*Timelapse)

	// Generate list of directories and timelapse files.
	for _, finfo := range files {
		rel, err := filepath.Rel(root, filepath.Join(b, finfo.Name()))
		if err != nil {
			return nil, err
		}
		if finfo.IsDir() {
			d := &Directory{
				Name: finfo.Name(),
				Path: rel,
			}
			r.Dirs = append(r.Dirs, d)
			continue
		}
		ms := timelapseRE.FindStringSubmatch(finfo.Name())
		if ms == nil || len(ms) != 4 {
			continue
		}
		prefix := ms[1]
		numStr := ms[2]
		ext := ms[3]

		var matchEXT bool
		for _, valid := range allowedEXT {
			if strings.ToLower(ext) == valid {
				matchEXT = true
			}
		}
		if !matchEXT {
			continue
		}

		num, err := strconv.Atoi(numStr)
		if err != nil {
			return nil, err
		}

		var t *Timelapse
		if lookup, ok := tmap[prefix]; ok {
			t = lookup
		} else {
			t = &Timelapse{
				Name:    finfo.Name(),
				Path:    rel,
				Prefix:  prefix,
				Ext:     ext,
				Start:   num,
				NumLen:  len(numStr),
				browser: f,
			}
			tmap[prefix] = t
		}

		// Increment counter if this is the next image in the sequence.
		if num == t.Start+t.Count {
			t.Count += 1
		}
	}

	// Extract timelapse map to list.
	for _, v := range tmap {
		fps := 60
		dur := time.Second * time.Duration(v.Count) / time.Duration(fps)
		v.DurationString = dur.Truncate(100 * time.Millisecond).String()

		r.Timelapses = append(r.Timelapses, v)
	}

	// Generates list of parents
	rel, err := filepath.Rel(root, b)
	if err != nil {
		return nil, err
	}
	pl := strings.Split(rel, string(os.PathSeparator))
	r.Parents = append(r.Parents, &Directory{
		Name: "[top]",
		Path: ".",
	})
	for i := range pl {
		d := &Directory{
			Name: pl[i],
			Path: filepath.Join(pl[:i+1]...),
		}
		if d.Path != "." {
			r.Parents = append(r.Parents, d)
		}
	}

	return r, nil
}

func (f *FileBrowser) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	p := r.Form.Get("path")
	response, err := f.listPath(p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	js, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

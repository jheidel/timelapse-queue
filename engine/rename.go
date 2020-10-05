package engine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"timelapse-queue/filebrowse"

	log "github.com/sirupsen/logrus"
)

func rename(src, dst string) error {
	// Avoid overwrites.
	_, err := os.Stat(dst)
	if !os.IsNotExist(err) {
		return fmt.Errorf("Output file %s already exists", dst)
	}
	// Do move.
	if err := os.Rename(src, dst); err != nil {
		return fmt.Errorf("Failed to move %s to %s: %v", src, dst, err)
	}
	return nil
}

func ConvertRename(ctx context.Context, logger *log.Logger, config Config, timelapse filebrowse.ITimelapse, progress chan<- int) error {

	start, end := config.GetStartEnd()
	skip := config.GetSkip()

	total := timelapse.ImageCount()

	i := 0
	for src := range filebrowse.ImagePaths(timelapse, start, end, skip) {
		ext := filepath.Ext(src)
		dst := timelapse.GetOutputFullPath(fmt.Sprintf("%s%06d%s", config.GetFilename(), i, ext))

		logger.Infof("Rename %q to %q", src, dst)
		if err := rename(src, dst); err != nil {
			logger.Errorf("FAILED: %v", err)
			return err
		}

		i++
		progress <- 100 * i / total
	}

	return nil
}

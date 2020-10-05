package engine

import (
	"context"
	"os"

	"timelapse-queue/filebrowse"

	"github.com/pkg/profile"
	log "github.com/sirupsen/logrus"
)

type ConvertOptions struct {
	ProfileCPU, ProfileMem bool
	Stack                  bool
	StackWindow            int
	StackSkipCount         int
	StackMode              string
	RenameOnly             bool
}

func Convert(ctx context.Context, config Config, timelapse filebrowse.ITimelapse, progress chan<- int) error {
	defer close(progress)
	opts := config.GetConvertOptions()

	profilepath := profile.ProfilePath(timelapse.GetOutputFullPath("profiles"))
	if opts.ProfileCPU {
		defer profile.Start(profilepath).Stop()
	}
	if opts.ProfileMem {
		defer profile.Start(profile.MemProfile, profilepath).Stop()
	}

	logf, err := os.Create(timelapse.GetOutputFullPath(config.GetDebugFilename()))
	if err != nil {
		return err
	}
	defer logf.Close()

	customFormatter := new(log.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	customFormatter.FullTimestamp = true

	logger := &log.Logger{
		Out:       logf,
		Formatter: customFormatter,
		Level:     log.DebugLevel,
	}

	if opts.RenameOnly {
		return ConvertRename(ctx, logger, config, timelapse, progress)
	}
	return ConvertFFMpeg(ctx, logger, config, timelapse, progress)
}

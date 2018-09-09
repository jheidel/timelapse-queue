package engine

import (
	"bufio"
	"context"
	"fmt"
	"image"
	"os"
	"os/exec"
	"regexp"
	"strconv"

	"timelapse-queue/filebrowse"
	"timelapse-queue/process"
	"timelapse-queue/util"

	"github.com/pkg/profile"

	log "github.com/sirupsen/logrus"
)

var (
	progressRE = regexp.MustCompile(`frame=\s*(\d+)`)
)

type ConvertOptions struct {
	ProfileCPU, ProfileMem bool
	Stack                  bool
	StackWindow            int
}

func Convert(pctx context.Context, config Config, timelapse *filebrowse.Timelapse, progress chan<- int) error {
	defer close(progress)
	opts := config.GetConvertOptions()

	profilepath := profile.ProfilePath(timelapse.GetOutputFullPath("profiles"))
	if opts.ProfileCPU {
		defer profile.Start(profilepath).Stop()
	}
	if opts.ProfileMem {
		defer profile.Start(profile.MemProfile, profilepath).Stop()
	}

	ctx, cancelf := context.WithCancel(pctx)

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

	// TODO: maybe use a filter chain in config to apply this sort of logic.
	start, end := config.GetStartEnd()
	imagec, imerrc := timelapse.Images(ctx, start, end)

	cropper := process.Crop{
		Region: config.GetRegion(),
	}
	imagec, imerrc = cropper.Process(imagec, imerrc)

	resizer := process.Resizer{
		Size: image.Point{X: 1920, Y: 1080},
	}
	imagec, imerrc = resizer.Process(imagec, imerrc)

	if opts.Stack {
		stacker := process.Stacker{
			Overlap: opts.StackWindow,
			Merger:  &process.Lighten{},
		}
		imagec, imerrc = stacker.Process(imagec, imerrc)
	}

	// Pull a sample image from the stream (to build config)
	var sample *image.RGBA
	select {
	case sample = <-imagec:
		break
	case err := <-imerrc:
		logger.Errorf("Failed to fetch a sample image: %v", err)
		return err
	}

	// Watch for future errors on the input channel. Any errors here should abort
	// ffmpeg (by context cancelation).
	go func() {
		err := <-imerrc
		if err != nil {
			logger.Errorf("Error reading image: %v", err)
			cancelf() // Abort ffmpeg.
		}
	}()

	args := []string{
		"-framerate", "60",

		"-f", "rawvideo",
		"-pixel_format", "bgr32",
		"-video_size", fmt.Sprintf("%dx%d", sample.Rect.Max.X, sample.Rect.Max.Y),
		"-i", "-", // Read from stdin.

		"-c:v", "libx264",
		"-preset", "slow",
		"-crf", "17",
		"-s", fmt.Sprintf("%dx%d", sample.Rect.Max.X, sample.Rect.Max.Y),
		"-progress", "/dev/stdout",
		timelapse.GetOutputFullPath(config.GetFilename()),
	}

	cmd := exec.Command(util.LocateFFmpegOrDie(), args...)
	r, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	stderr := bufio.NewScanner(r)

	go func() {
		for stderr.Scan() {
			logger.Error(stderr.Text())
		}
	}()

	r, err = cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stdout := bufio.NewScanner(r)
	go func() {
		for stdout.Scan() {
			l := stdout.Text()
			logger.Info(l)

			m := progressRE.FindStringSubmatch(l)
			if len(m) != 2 {
				continue
			}
			i, err := strconv.Atoi(m[1])
			if err != nil {
				log.Errorf("Failed to convert frame number %s to int", m[1])
				continue
			}
			progress <- 100 * i / config.GetExpectedFrames()
		}
	}()

	// Stream images to ffmpeg.
	pin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	go func() {
		defer pin.Close()
		// Make sure to include the sample image we took earlier.
		_, err := pin.Write(sample.Pix)
		if err != nil {
			logger.Errorf("Failed to write initial image to ffmpeg: %v", err)
			cancelf()
			return
		}
		// Write remainder of image stream to ffmpeg.
		for img := range imagec {
			_, err := pin.Write(img.Pix)
			if err != nil {
				logger.Errorf("Failed to write image to ffmpeg: %v", err)
				cancelf()
				return
			}
		}
	}()

	logger.Infof("Running FFmpeg with args: %v", args)

	if err := cmd.Start(); err != nil {
		log.Errorf("Failed to start FFmpeg subprocess")
		return err
	}

	errc := make(chan error)
	go func() {
		errc <- cmd.Wait()
	}()

	donec := ctx.Done()

	for {
		select {
		case err := <-errc:
			if err != nil {
				log.Warnf("Conversion failed: %v.", err)
				return err
			}
			log.Info("Conversion succeeded.")
			return nil
		case <-donec:
			donec = nil
			log.Warnf("Context cancel (%v), aborting FFmpeg", ctx.Err())
			logger.Warnf("Context cancel (%v), aborting FFmpeg", ctx.Err())
			if err := cmd.Process.Signal(os.Interrupt); err != nil {
				log.Infof("Failed to signal FFmpeg for context cancel: %v", err)
				return err
			}
		}
	}
}

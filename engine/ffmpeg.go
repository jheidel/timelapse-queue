package engine

import (
	"bufio"
	"context"
	"fmt"
	"image"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"time"

	"timelapse-queue/filebrowse"
	"timelapse-queue/process"
	"timelapse-queue/util"

	"github.com/pkg/profile"

	log "github.com/sirupsen/logrus"
)

var (
	progressRE = regexp.MustCompile(`frame=\s*(\d+)`)
)

const (
	watchdogDuration = 5 * time.Minute
	frameDeadline    = 4 * time.Minute
)

type ConvertOptions struct {
	ProfileCPU, ProfileMem bool
	Stack                  bool
	StackWindow            int
	StackSkipCount         int
	StackMode              string
}

// imageWriter writes RGBA images directly to FFmpeg to be used as rawvideo input.
type imageWriter struct {
	out    io.Writer
	bufOut *bufio.Writer
}

func newImageWriter(w io.Writer) *imageWriter {
	return &imageWriter{
		out: w,
	}
}

func (w *imageWriter) Write(img *image.RGBA) error {
	sz := img.Rect.Dx() * img.Rect.Dy() * 4
	if len(img.Pix) == sz {
		// Region covers the entire pixel buffer, simply write everything directly to output.
		if _, err := w.out.Write(img.Pix); err != nil {
			return err
		}
		return nil
	}

	// Need to write out one stride at a time.
	// Each frame needs to be complete though for FFmpeg, so use an output buffer.
	if w.bufOut == nil {
		w.bufOut = bufio.NewWriterSize(w.out, sz)
	}
	if w.bufOut.Available() < sz {
		return fmt.Errorf("buffer size unexpectedly too small, want %v got %v", sz, w.bufOut.Available())
	}
	// TODO(jheidel): the docs seem wrong here since this should be base on img.Rect.Min but it isn't.
	p := 0 // pix.Buf starts at the origin
	for i := 0; i < img.Rect.Dy(); i++ {
		rowEnd := p + img.Rect.Dx()*4
		if rowEnd > len(img.Pix) {
			return fmt.Errorf("unexpected end of pix buf from position %d, size %d", p, len(img.Pix))
		}
		if _, err := w.bufOut.Write(img.Pix[p:rowEnd]); err != nil {
			return err
		}
		p += img.Stride
	}
	if err := w.bufOut.Flush(); err != nil {
		return err
	}

	return nil
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
	defer cancelf()

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

	outp, err := config.GetOutputProfile()
	if err != nil {
		return err
	}

	// TODO: maybe use a filter chain in config to apply this sort of logic.
	start, end := config.GetStartEnd()
	imagec, imerrc := timelapse.Images(ctx, start, end)

	cropper := process.Crop{
		Region: config.GetRegion(),
	}
	imagec, imerrc = cropper.Process(imagec, imerrc)

	resizer := process.Resizer{
		Size: image.Point{X: outp.Width, Y: outp.Height},
	}
	imagec, imerrc = resizer.Process(imagec, imerrc)

	if opts.Stack {
		stacker := process.Stacker{
			Overlap: opts.StackWindow,
			Skip:    opts.StackSkipCount,
			Merger:  process.GetMergerByName(opts.StackMode),
		}
		imagec, imerrc = stacker.Process(imagec, imerrc)
	}

	// Writes errors both to the system logger and the file logger.
	dualErrorf := func(format string, v ...interface{}) {
		log.Errorf(format, v...)
		logger.Errorf(format, v...)
	}

	// Pull a sample image from the stream (to build config)
	var sample *image.RGBA
	sampleDeadline := time.NewTimer(frameDeadline)
	select {
	case sample = <-imagec:
		break
	case err := <-imerrc:
		dualErrorf("Failed to fetch a sample image: %v", err)
		return err
	case <-sampleDeadline.C:
		dualErrorf("Deadline exceeded fetching sample image")
		return fmt.Errorf("deadline exceeded fetching sample")
	}

	// Watch for future errors on the input channel. Any errors here should abort
	// ffmpeg (by context cancelation).
	go func() {
		err := <-imerrc
		if err != nil {
			dualErrorf("Error reading image: %v", err)
			cancelf() // Abort ffmpeg.
		}
	}()

	// Monitor for cases where FFmpeg becomes non-responsive (stops sending
	// status updates). If so, abort ffmpeg to allow the queue to progress.
	watchdog := time.NewTimer(watchdogDuration)
	go func() {
		select {
		case <-watchdog.C:
			dualErrorf("Watchdog expiration: FFmpeg non-responsive after %v, cancelling", watchdogDuration)
			cancelf()
		case <-ctx.Done():
		}
	}()

	args := []string{
		"-framerate", fmt.Sprintf("%d", config.GetFPS()),
		"-f", "rawvideo",
		"-pixel_format", "bgr32",
		"-video_size", fmt.Sprintf("%dx%d", sample.Rect.Dx(), sample.Rect.Dy()),
		"-i", "-", // Read from stdin.

		"-c:v", "libx264",
		"-preset", "slow",
		"-crf", "16",
	}
	args = append(args, outp.FFmpegArgs...)
	args = append(args, []string{
		"-x264opts", "colorprim=bt709:transfer=bt709:colormatrix=bt709:fullrange=off",
		"-s", fmt.Sprintf("%dx%d", sample.Rect.Dx(), sample.Rect.Dy()),
		"-progress", "/dev/stdout",
		timelapse.GetOutputFullPath(config.GetFilename()),
	}...)

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
			watchdog.Reset(watchdogDuration) // pet
		}
	}()

	// Stream images to ffmpeg.
	pin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	go func() {
		defer pin.Close()
		imgWriter := newImageWriter(pin)
		// Make sure to include the sample image we took earlier.
		if err := imgWriter.Write(sample); err != nil {
			dualErrorf("Failed to write initial image to ffmpeg: %v", err)
			cancelf()
			return
		}

		for {
			deadline := time.NewTimer(frameDeadline)
			select {
			case <-deadline.C:
				dualErrorf("Deadline exceeded waiting for next frame")
				cancelf()
				return
			case img, ok := <-imagec:
				if !ok {
					// Done with image sequence.
					return
				}
				if err := imgWriter.Write(img); err != nil {
					dualErrorf("Failed to write image to ffmpeg: %v", err)
					cancelf()
					return
				}
			}
		}
	}()

	logger.Infof("Running FFmpeg with args: %v", args)

	if err := cmd.Start(); err != nil {
		dualErrorf("Failed to start FFmpeg subprocess")
		return err
	}

	errc := make(chan error)
	go func() {
		errc <- cmd.Wait()
	}()

	donec := ctx.Done()
	var killc <-chan time.Time

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
			killc = time.After(2 * time.Minute)
		case <-killc:
			log.Warnf("FFmpeg cancel taking too long, sending SIGKILL")
			logger.Warnf("FFmpeg cancel taking too long, sending SIGKILL")
			if err := cmd.Process.Signal(os.Kill); err != nil {
				log.Infof("Failed to SIGKILL FFmpeg: %v", err)
				return err
			}
		}
	}
}

package engine

import (
	"bufio"
	"context"
	"os"
	"os/exec"
	"regexp"
	"strconv"

	"timelapse-queue/filebrowse"
	"timelapse-queue/util"

	log "github.com/sirupsen/logrus"
)

var (
	progressRE = regexp.MustCompile(`frame=\s*(\d+)`)
)

func Convert(ctx context.Context, config Config, timelapse *filebrowse.Timelapse, progress chan<- int) error {
	defer close(progress)

	args := config.GetArgs(timelapse)
	cmd := exec.Command(util.LocateFFmpegOrDie(), args...)

	logf, err := os.Create(config.GetDebugFullPath(timelapse))
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

	r, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	stderr := bufio.NewScanner(r)

	go func() {
		for stderr.Scan() {
			l := stderr.Text()
			logger.Error(l)

			m := progressRE.FindStringSubmatch(l)
			if len(m) != 2 {
				continue
			}
			i, err := strconv.Atoi(m[1])
			if err != nil {
				log.Errorf("Failed to convert frame number %s to int", m[1])
				continue
			}
			progress <- 100 * i / timelapse.Count
		}
	}()

	r, err = cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stdout := bufio.NewScanner(r)
	go func() {
		for stdout.Scan() {
			logger.Info(stdout.Text())
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

package engine

import (
	"bufio"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"

	"timelapse-queue/util"

	log "github.com/sirupsen/logrus"
)

var (
	progressRE = regexp.MustCompile(`frame=\s*(\d+)`)
)

type FFmpegConfig struct {
}

func Convert() error {
	args := []string{
		"-r", "60",
		"-start_number", "21015",
		"-i", "/home/jeff/timelapse/G%07d.JPG",
		"-c:v", "libx264",
		"-preset", "slow",
		"-crf", "17",
		"-s", "1920x1080",
		"/home/jeff/timelapse/1080p-test.mp4",
	}
	cmd := exec.Command(util.LocateFFmpegOrDie(), args...)

	logf, err := os.Create("/home/jeff/timelapse/1080p-test.mp4.log")
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

			log.Infof("Currently on frame %d", i)
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

	if err := cmd.Run(); err != nil {
		return err
	}

	log.Infof("Successful convert.")

	return nil
}

type TestServer struct {
}

func (s *TestServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := Convert()
	if err != nil {
		log.Errorf("Convert returned error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("done"))
}

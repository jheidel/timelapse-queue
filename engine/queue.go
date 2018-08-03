package engine

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"timelapse-queue/filebrowse"

	log "github.com/sirupsen/logrus"
)

type jobState string

const (
	StatePending = jobState("pending")
	StateActive  = jobState("active")
	StateDone    = jobState("done")
	StateFailed  = jobState("failed")
)

type Job struct {
	State     jobState
	ID        int
	LogPath   string
	Progress  int
	Timelapse *filebrowse.Timelapse
	Config    Config

	// Elapsed time, derived from start time. Updated as part of JSON serialization.
	ElapsedString string
	start         time.Time
	stop          time.Time
}

type JobQueue struct {
	Queue []*Job

	jobIDgen int

	current      *Job
	jobdonec     chan error
	jobprogressc chan int

	serialc chan chan *jsonResp
	addc    chan *Job
}

func NewJobQueue() *JobQueue {
	return &JobQueue{
		Queue:   []*Job{},
		serialc: make(chan chan *jsonResp),
		addc:    make(chan *Job),
	}
}

func (q *JobQueue) nextJob() *Job {
	for _, j := range q.Queue {
		if j.State == StatePending {
			return j
		}
	}
	return nil
}

func (q *JobQueue) maybeStartNext(ctx context.Context) {
	if q.current != nil {
		return // Job already running.
	}
	j := q.nextJob()
	if j == nil {
		return // No jobs remaining.
	}
	// Start next job.
	q.jobdonec = make(chan error)
	q.jobprogressc = make(chan int)
	q.current = j
	j.State = StateActive
	j.LogPath = j.Config.GetDebugPath(j.Timelapse)
	j.start = time.Now()
	go func() {
		defer close(q.jobdonec)
		q.jobdonec <- Convert(ctx, j.Config, j.Timelapse, q.jobprogressc)
	}()
	log.Infof("job started")
}

func (q *JobQueue) markJobDone(err error) {
	j := q.current
	j.stop = time.Now()
	if err != nil {
		j.State = StateFailed
		j.Progress = 0
	} else {
		j.State = StateDone
		j.Progress = 100
	}

	q.jobdonec = nil
	q.jobprogressc = nil
	q.current = nil

	log.Infof("job completed")
}

func (q *JobQueue) AddJob(config Config, t *filebrowse.Timelapse) {
	j := &Job{
		State:     StatePending,
		Timelapse: t,
		Config:    config,
		ID:        q.jobIDgen,
	}
	q.jobIDgen += 1
	q.addc <- j
}

func (q *JobQueue) toJSON() *jsonResp {
	now := time.Now()
	for _, j := range q.Queue {
		end := now
		if !j.stop.IsZero() {
			end = j.stop
		}
		if !j.start.IsZero() {
			j.ElapsedString = end.Sub(j.start).Truncate(time.Second).String()
		}
	}

	r, err := json.Marshal(q)
	return &jsonResp{
		Result: r,
		Err:    err,
	}
}

func (q *JobQueue) Loop(ctx context.Context) {
	exitc := ctx.Done()
	log.Info("starting job queue")

	for {
		select {
		case j := <-q.addc:
			q.Queue = append(q.Queue, j)
			log.Info("new job added to queue")
			q.maybeStartNext(ctx)
		case err := <-q.jobdonec:
			q.markJobDone(err)
			q.maybeStartNext(ctx)
		case p := <-q.jobprogressc:
			q.current.Progress = p
		case c := <-q.serialc:
			c <- q.toJSON()
		case <-exitc:
			log.Info("job queue stopping")
			return
		}
	}
}

type jsonResp struct {
	Result []byte
	Err    error
}

func (q *JobQueue) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := make(chan *jsonResp)
	q.serialc <- c
	resp := <-c
	if resp.Err != nil {
		http.Error(w, resp.Err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp.Result)
}

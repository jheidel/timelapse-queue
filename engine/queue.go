package engine

import (
	"context"
	"encoding/json"
	"net/http"

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
	Progress  int
	Timelapse *filebrowse.Timelapse
	Config    Config
}

type JobQueue struct {
	Queue []*Job

	current      *Job
	jobdonec     chan error
	jobprogressc chan int

	serialc chan chan *jsonResp
	addc    chan *Job
}

func (q *JobQueue) NextJob() *Job {
	for _, j := range q.Queue {
		if j.State == StatePending {
			return j
		}
	}
	return nil
}

func NewJobQueue() *JobQueue {
	return &JobQueue{
		Queue:   []*Job{},
		serialc: make(chan chan *jsonResp),
		addc:    make(chan *Job),
	}
}

func (q *JobQueue) maybeStartNext(ctx context.Context) {
	if q.current != nil {
		return // Job already running.
	}
	j := q.NextJob()
	if j == nil {
		return // No jobs remaining.
	}
	// Start next job.
	q.jobdonec = make(chan error)
	q.jobprogressc = make(chan int)
	q.current = j
	j.State = StateActive
	go func() {
		defer close(q.jobdonec)
		q.jobdonec <- Convert(ctx, j.Config, j.Timelapse, q.jobprogressc)
	}()
	log.Infof("job started")
}

func (q *JobQueue) markJobDone(err error) {
	j := q.current
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
	}
	q.addc <- j
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
			r, err := json.Marshal(q)
			resp := &jsonResp{
				Result: r,
				Err:    err,
			}
			c <- resp
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

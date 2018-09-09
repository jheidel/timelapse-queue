package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"timelapse-queue/filebrowse"

	log "github.com/sirupsen/logrus"
)

type jobState string

const (
	StatePending = jobState("pending")
	StateActive  = jobState("active")
	StateDone    = jobState("done")
	StateCancel  = jobState("cancel")
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

	// Cancels this job.
	cancelf context.CancelFunc
}

type JobQueue struct {
	Queue []*Job

	jobIDgen int

	current      *Job
	jobdonec     chan error
	jobprogressc chan int

	serialc chan chan *jsonResp
	addc    chan *Job
	cancelc chan *jobOp
	removec chan *jobOp
}

func NewJobQueue() *JobQueue {
	return &JobQueue{
		Queue:   []*Job{},
		serialc: make(chan chan *jsonResp),
		addc:    make(chan *Job),
		cancelc: make(chan *jobOp),
		removec: make(chan *jobOp),
	}
}

type jobOp struct {
	ID   int
	Errc chan error
}

func (q *JobQueue) nextJob() *Job {
	for _, j := range q.Queue {
		if j.State == StatePending {
			return j
		}
	}
	return nil
}

func (q *JobQueue) getJob(ID int) *Job {
	for _, j := range q.Queue {
		if j.ID == ID {
			return j
		}
	}
	return nil
}

func (q *JobQueue) removeJob(ID int) error {
	for i, j := range q.Queue {
		if j.ID == ID {
			if j.State == StateActive {
				return fmt.Errorf("could not remove job %v in state %v", ID, j.State)
			}
			q.Queue = append(q.Queue[0:i], q.Queue[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("job %v not found in queue", ID)
}

func (q *JobQueue) cancelJob(ID int) error {
	j := q.getJob(ID)
	if j == nil {
		return fmt.Errorf("job %v not found", ID)
	}

	if j.State != StateActive {
		return fmt.Errorf("job not in active state and cannot be canceled")
	}
	if j.cancelf == nil {
		return fmt.Errorf("job not running or already canceled")
	}

	j.cancelf()
	j.cancelf = nil
	j.State = StateCancel
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
	j.State = StateActive
	j.LogPath = j.Timelapse.GetOutputPath(j.Config.GetDebugFilename())
	j.start = time.Now()

	jobCtx, cancel := context.WithCancel(ctx)
	q.jobdonec = make(chan error)
	q.jobprogressc = make(chan int)
	j.cancelf = cancel
	go func() {
		defer close(q.jobdonec)
		q.jobdonec <- Convert(jobCtx, j.Config, j.Timelapse, q.jobprogressc)
	}()
	q.current = j
	log.Infof("job started")
}

func (q *JobQueue) markJobDone(err error) {
	j := q.current
	j.stop = time.Now()
	if err != nil {
		j.State = StateFailed
	} else {
		j.State = StateDone
		j.Progress = 100
	}

	q.jobdonec = nil
	q.jobprogressc = nil
	q.current = nil

	log.Infof("job completed")

	// Ensure we run a GC cycle before running the next job.
	// There are probably heap fragmentation issues that are causing more problems...
	runtime.GC()
	log.Infof("gc complete")
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
		case t := <-q.cancelc:
			log.Infof("issue cancel of job %d", t.ID)
			err := q.cancelJob(t.ID)
			t.Errc <- err
		case t := <-q.removec:
			log.Infof("remove job %d", t.ID)
			err := q.removeJob(t.ID)
			t.Errc <- err
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

func (q *JobQueue) handlePostJobOp(w http.ResponseWriter, r *http.Request, opc chan *jobOp) {
	if r.Method != "POST" {
		http.Error(w, "Requires POST", http.StatusBadRequest)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ID, err := strconv.Atoi(r.Form.Get("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	t := &jobOp{
		ID:   ID,
		Errc: make(chan error),
	}
	opc <- t
	err = <-t.Errc
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (q *JobQueue) ServeCancel(w http.ResponseWriter, r *http.Request) {
	q.handlePostJobOp(w, r, q.cancelc)
}

func (q *JobQueue) ServeRemove(w http.ResponseWriter, r *http.Request) {
	q.handlePostJobOp(w, r, q.removec)
}

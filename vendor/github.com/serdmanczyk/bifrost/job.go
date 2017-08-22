package bifrost

import (
	"encoding/json"
	"sync"
	"time"
)

// JobRunner interface is the accepted type for a job to be dispatched in the
// background.
type JobRunner interface {
	Run() error
}

// JobRunnerFunc is a type used to wrap pure functions,
// allowing them to implement the JobRunner interface.
type JobRunnerFunc func() error

// Run implements the JobRunner interface for a JobRunnerFunc.
func (f JobRunnerFunc) Run() error {
	return f()
}

// JobTracker is the interface made available to query a job's status
type JobTracker interface {
	ID() uint
	Done() <-chan bool
	Status() JobStatus
}

// JobTrackers holds an array of JobTrackers
type JobTrackers []JobTracker

// MarshalJSON allows an array of JobTrackers to marshal their statuses
func (ts JobTrackers) MarshalJSON() ([]byte, error) {
	js := make([]JobStatus, 0)

	for _, t := range ts {
		js = append(js, t.Status())
	}

	return json.Marshal(js)
}

func newJob(j JobRunner, id uint) *job {
	job := &job{
		id:       id,
		start:    time.Now(),
		done:     make(chan bool, 1),
		complete: false,
		success:  false,
	}

	var once sync.Once
	job.run = func() {
		once.Do(func() {
			err := j.Run()

			job.lock.Lock()
			defer job.lock.Unlock()

			job.finish = time.Now()
			job.complete = true
			job.success = err == nil
			job.err = err
			job.done <- true
			close(job.done)
		})
	}

	return job
}

type job struct {
	id       uint
	lock     sync.RWMutex
	complete bool
	success  bool
	err      error
	start    time.Time
	finish   time.Time
	done     chan bool
	run      func()
}

// JobStatus represents the status of a job at a given instant in time.
type JobStatus struct {
	ID       uint
	Complete bool
	Success  bool
	Error    string
	Start    time.Time
	Finish   time.Time
}

// helper struct for (j JobStatus) MarshalJSON()
type jobStatusTransport struct {
	ID       uint
	Complete bool
	Success  *bool  `json:",omitempty"`
	Error    string `json:",omitempty"`
	Start    time.Time
	Finish   *time.Time `json:",omitempty"`
}

// MarshalJSON is implemented to allow custom inclusion/exclusion of
// 'zero value' job status fields.
func (j JobStatus) MarshalJSON() ([]byte, error) {

	// ignore finish if not set yet
	var finish *time.Time
	var zeroTime time.Time
	if j.Finish != zeroTime {
		finish = &j.Finish
	}

	// ignore success if not complete
	var success *bool
	if j.Complete {
		success = &j.Success
	}

	return json.Marshal(jobStatusTransport{
		ID:       j.ID,
		Complete: j.Complete,
		Success:  success,
		Error:    j.Error,
		Start:    j.Start,
		Finish:   finish,
	})
}

// ID returns the job's ID
func (j *job) ID() uint {
	return j.id
}

// Done returns a channel that can be used to wait on job completion
func (j *job) Done() <-chan bool {
	return j.done
}

// Status returns the job's current status as a JobStatus struct.
func (j *job) Status() JobStatus {
	j.lock.RLock()
	defer j.lock.RUnlock()

	var errorMessage string
	if j.err != nil {
		errorMessage = j.err.Error()
	}

	return JobStatus{
		ID:       j.id,
		Complete: j.complete,
		Success:  j.success,
		Error:    errorMessage,
		Start:    j.start,
		Finish:   j.finish,
	}
}

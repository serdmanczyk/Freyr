package bifrost

import (
	"fmt"
	"sync"
	"time"
)

const (
	defaultNumWorkers int           = 4
	defaultJobExpiry  time.Duration = time.Minute * 5
	jobChanMargin     int           = 2
	jobMapCapMargin   int           = 5
)

// JobDispatcher defines an interface for dispatching jobs
type JobDispatcher interface {
	JobStatus(jobID uint) (JobTracker, error)
	Jobs() JobTrackers
	Queue(j JobRunner) JobTracker
	QueueFunc(j JobRunnerFunc) JobTracker
}

// WorkerDispatcher is used to maintain and delegate jobs to workers.
type WorkerDispatcher struct {
	workerQueue chan chan *job
	jobchan     chan *job
	stop        chan bool
	numWorkers  int
	currID      uint
	jobExpiry   time.Duration
	workers     []*worker
	jobsMap     map[uint]JobTracker
	mapLock     sync.RWMutex
}

// NewDispatcher create a new Dispatcher and initializes its workers.
func NewWorkerDispatcher(opts ...WorkerDispatcherOpt) *WorkerDispatcher {
	d := &WorkerDispatcher{
		numWorkers:  defaultNumWorkers,
		jobExpiry:   defaultJobExpiry,
		workerQueue: make(chan chan *job),
		stop:        make(chan bool),
	}

	for _, option := range opts {
		option(d)
	}

	d.workers = make([]*worker, 0, d.numWorkers)
	d.jobchan = make(chan *job, d.numWorkers*jobChanMargin)
	d.jobsMap = make(map[uint]JobTracker, d.numWorkers*jobMapCapMargin)

	for i := 0; i < d.numWorkers; i++ {
		worker := newWorker(d.workerQueue)
		worker.start()
		d.workers = append(d.workers, worker)
	}
	d.start()
	return d
}

func (d *WorkerDispatcher) start() {
	auditTicker := time.NewTicker(d.jobExpiry)
	go func() {
		for {
			select {
			case j := <-d.jobchan:
				go func(j *job) {
					worker := <-d.workerQueue
					worker <- j
				}(j)
			case <-auditTicker.C:
				go d.jobMapAudit()
			case <-d.stop:
				d.stopWorkers()
				d.stop <- true
				close(d.stop)
				return
			}
		}
	}()
}

func (d *WorkerDispatcher) stopWorkers() {
	stopChans := make([]chan bool, 0, len(d.workers))

	// signal all workers to finish what they're doing
	for _, worker := range d.workers {
		worker.stop <- true
		stopChans = append(stopChans, worker.stop)
	}
	// wait for workers to finish
	for _, stopChan := range stopChans {
		<-stopChan
	}
}

// Stop signals all workers to stop running their current
// jobs, waits for them to finish, then returns.
func (d *WorkerDispatcher) Stop() {
	d.stop <- true
	<-d.stop
}

func (d *WorkerDispatcher) jobMapAudit() {
	remIDs := make([]uint, 0, jobMapCapMargin)

	d.mapLock.RLock()
	for id, job := range d.jobsMap {
		status := job.Status()

		if status.Complete && time.Now().Sub(status.Finish) > d.jobExpiry {
			remIDs = append(remIDs, id)
		}
	}
	d.mapLock.RUnlock()

	if len(remIDs) > 0 {
		d.mapLock.Lock()
		for _, id := range remIDs {
			delete(d.jobsMap, id)
		}
		d.mapLock.Unlock()
	}
}

// Queue takes an implementer of the JobRunner interface and schedules it to
// be run via a worker.
func (d *WorkerDispatcher) Queue(j JobRunner) JobTracker {
	d.mapLock.Lock()
	defer d.mapLock.Unlock()

	job := newJob(j, d.currID)
	d.jobsMap[job.id] = job
	d.currID++
	d.jobchan <- job

	return job
}

// QueueFunc is a convenience function for queuing a JobRunnerFunc
func (d *WorkerDispatcher) QueueFunc(j JobRunnerFunc) JobTracker {
	return d.Queue(JobRunner(j))
}

// Jobs returns all currently managed jobs
func (d *WorkerDispatcher) Jobs() JobTrackers {
	d.mapLock.RLock()
	defer d.mapLock.RUnlock()

	var trackers JobTrackers
	for _, job := range d.jobsMap {
		trackers = append(trackers, job)
	}

	return trackers
}

// JobStatus returns a JobTracker for the given jobID.  The JobTracker interface
// can be used to query status.
func (d *WorkerDispatcher) JobStatus(jobID uint) (JobTracker, error) {
	d.mapLock.RLock()
	defer d.mapLock.RUnlock()

	job, ok := d.jobsMap[jobID]
	if !ok {
		return nil, fmt.Errorf("Job %d not found", jobID)
	}

	return job, nil
}

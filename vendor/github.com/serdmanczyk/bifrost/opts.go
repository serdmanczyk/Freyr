package bifrost

import (
	"time"
)

// WorkerDispatcherOpt is a function type for
// configuring new WorkerDispatchers
type WorkerDispatcherOpt func(*WorkerDispatcher)

// Workers sets the number of workers to spawn
func Workers(numWorkers int) WorkerDispatcherOpt {
	return func(d *WorkerDispatcher) {
		d.numWorkers = numWorkers
	}
}

// JobExpiry sets the duration after a job completes
// to maintain the job's info for querying.  After
// expiry elapses, the job info will be purged.
//	Note: JobExpiry only controls the purge of FINISHED jobs.
//	There is currently not a provision for stopping a running
//	job that has continued to run beyond its expected duration.
func JobExpiry(expiry time.Duration) WorkerDispatcherOpt {
	return func(d *WorkerDispatcher) {
		d.jobExpiry = expiry
	}
}

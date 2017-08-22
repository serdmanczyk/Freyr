// Package bifrost contains functionality to create an in process job queue with
// a configurable number of workers running goroutines.  It also includes the
// ability to query scheduled jobs for status (completed jobs are purged at a
// configurable interval)
package bifrost

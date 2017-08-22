package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/serdmanczyk/bifrost"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"
)

const (
	defaultNumWorkers     int           = 10
	defaultNumJobs        int           = 10000
	defaultJobDuration    time.Duration = time.Microsecond
	defaultJobExpiration  time.Duration = time.Minute * 5
	defaultReportInterval time.Duration = time.Millisecond * 200
)

var (
	numWorkers     = flag.Int("workers", defaultNumWorkers, "number of workers to spawn")
	numJobs        = flag.Int("jobs", defaultNumJobs, "number of jobs to create")
	jobDuration    = flag.Int64("jobduration", int64(defaultJobDuration), "How long jobs last (time.Sleep")
	jobExpiry      = flag.Int64("expiration", int64(defaultJobExpiration), "How long until a finished job is purged")
	report         = flag.Bool("report", false, "Report on random jobs while jobs are running")
	reportInterval = flag.Int64("reportinterval", int64(defaultReportInterval), "Interval on which to report a random job's status (if report enabled)")
)

func init() {
	rand.Seed(time.Now().Unix())
}

func main() {
	flag.Parse()
	elapsed := func() func() time.Duration {
		start := time.Now()
		return func() time.Duration {
			return time.Now().Sub(start)
		}
	}()

	var wg sync.WaitGroup
	jobIDs := make([]uint, 0, *numJobs)

	dispatcher := bifrost.NewWorkerDispatcher(
		bifrost.Workers(*numWorkers),
		bifrost.JobExpiry(time.Duration(*jobExpiry)),
	)

	log.Printf("initialized %s\n", elapsed())

	jobFunc := func() error {
		defer wg.Done()

		time.Sleep(time.Duration(*jobDuration))
		if rand.Float32() > 0.5 {
			return fmt.Errorf("The odds are not in your favor")
		}

		return nil
	}

	for i := 0; i < *numJobs; i++ {
		wg.Add(1)
		job := dispatcher.QueueFunc(jobFunc)
		id := job.ID()
		jobIDs = append(jobIDs, id)
	}
	log.Printf("queued %s\n", elapsed())

	done := make(chan bool)
	go func() {
		wg.Wait()
		dispatcher.Stop()
		done <- true
		close(done)
	}()

	for {
		select {
		case <-time.After(time.Duration(*reportInterval)):
			if *report {
				// Grab a random job and print its status
				id := jobIDs[rand.Intn(len(jobIDs))]
				job, err := dispatcher.JobStatus(id)
				if err != nil {
					// ignore, job may have been purged
					continue
				}
				status := job.Status()
				json.NewEncoder(os.Stdout).Encode(&status)
				<-job.Done()
			}
		case <-done:
			goto complete
		}
	}

complete:
	log.Printf("done %s\n", elapsed())
}

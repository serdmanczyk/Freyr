 ᛉ Bifröst - a queryable in-process worker queue
------
[![Go Report Card](https://goreportcard.com/badge/github.com/serdmanczyk/Bifrost)](https://goreportcard.com/report/github.com/serdmanczyk/Bifrost)
[![GoDoc](https://godoc.org/github.com/serdmanczyk/Bifrost?status.svg)](https://godoc.org/github.com/serdmanczyk/Bifrost)
[![blog](https://img.shields.io/badge/readMy-blog-green.svg)](http://serdmanczyk.github.io)

![gofrost](repo/gofrost.jpg "Freyr")

Package bifrost contains functionality to create an in-process job queue with a configurable number of goroutine via workers.  It also includes the ability to query scheduled jobs for status (completed jobs are purged at a configurable interval)

```golang
package main

import (
    "encoding/json"
    "fmt"
    "github.com/serdmanczyk/bifrost"
    "os"
    "time"
)

func main() {
    stdoutWriter := json.NewEncoder(os.Stdout)
    dispatcher := bifrost.NewDispatcher(
        bifrost.Workers(4),
        bifrost.JobExpiry(time.Millisecond),
    )

    // Queue a job func
    tracker := dispatcher.QueueFunc(func() error {
        time.Sleep(time.Microsecond)
        return nil
    })

    // Queue a 'JobRunner'
    dispatcher.Queue(bifrost.JobRunnerFunc(func() error {
        time.Sleep(time.Microsecond)
        return nil
    }))

    // Print out incomplete status
    status := tracker.Status()
    stdoutWriter.Encode(&status)
    // {"ID":0,"Complete":false,"Start":"2017-03-23T21:51:27.140681968-07:00"}

    // wait on completion
    <-tracker.Done()
    // Status is now complete
    status = tracker.Status()
    stdoutWriter.Encode(&status)
    // {"ID":0,"Complete":true,"Success":true,"Start":"2017-03-23T21:51:27.140681968-07:00","Finish":"2017-03-23T21:51:27.140830827-07:00"}

    // Queue a job that will 'fail'
    tracker = dispatcher.QueueFunc(func() error {
        time.Sleep(time.Microsecond)
        return fmt.Errorf("Failed")
    })

    // Show failure status
    <-tracker.Done()
    status = tracker.Status()
    stdoutWriter.Encode(&status)
    // {"ID":2,"Complete":true,"Success":false,"Error":"Failed","Start":"2017-03-23T21:51:27.141026625-07:00","Finish":"2017-03-23T21:51:27.141079871-07:00"}

    // Query for a job's status.
    tracker, _ = dispatcher.Status(tracker.ID())
    status = tracker.Status()
    stdoutWriter.Encode(&status)
    // {"ID":2,"Complete":true,"Success":false,"Error":"Failed","Start":"2017-03-23T21:51:27.141026625-07:00","Finish":"2017-03-23T21:51:27.141079871-07:00"}

    // Show all jobs
    jobs := dispatcher.GetJobs()
    stdoutWriter.Encode(jobs)
    // [{"ID":2,"Complete":true,"Success":false,"Error":"Failed","Start":"2017-03-23T21:51:27.141026625-07:00","Finish":"2017-03-23T21:51:27.141079871-07:00"},{"ID":0,"Complete":true,"Success":true,"Start":"2017-03-23T21:51:27.140681968-07:00","Finish":"2017-03-23T21:51:27.140830827-07:00"},{"ID":1,"Complete":true,"Success":true,"Start":"2017-03-23T21:51:27.140684331-07:00","Finish":"2017-03-23T21:51:27.140873087-07:00"}]

    // wait for jobs to be purged
    time.Sleep(time.Millisecond * 5)

    // should now be empty
    jobs = dispatcher.GetJobs()
    stdoutWriter.Encode(jobs)
    // []

    dispatcher.Stop()
}
```

## Why?

If you've read the blog posts [Handling 1 Million Requests per Minute with Go](http://marcio.io/2015/07/handling-1-million-requests-per-minute-with-golang/) or [Writing worker queues, in Go](http://nesv.github.io/golang/2014/02/25/worker-queues-in-go.html) this will look very familiar.  The main machinery in Bifrost is basically identical to the functionality described in those blog posts, but with a couple added features I wanted for my project.

Added Features:

- Generic jobs: any `func() error` or type that implements `func Run() error` can be queued as a job.
- Graceful shutdown: when dispatcher is stopped, waits for running jobs to complete.
- Tracking: queued jobs are given an ID that can be used to query for status later.
- Cleanup: completed jobs are cleaned up after a configurable amount of time.

Lacks (might add these later):

- Lost jobs: if the dispatcher is stopped before all jobs are sent to a worker, unsent jobs may be ignored.
- Errant jobs: jobs taking longer than expected cannot be cancelled.
- Single process: this package does not include functionality to schedule jobs across multiple processes via AMQP, gRPC, or otherwise.

For an example, see the [test](dispatcher_test.go) or example [command line app](example/main.go).

Obligatory "not for use in production" but I do welcome feedback.

## Etymology

[Bifröst](https://en.wikipedia.org/wiki/Bifr%C3%B6st) (pronounce B-eye-frost popularly, or traditionally more like Beefroast) is the bridge between the realms of Earth and Asgard (the heavens) in norse mythology.

The Futhark [ᛉ Elhaz/Algiz](https://en.wikipedia.org/wiki/Algiz) is seen as the symbol for Bifröst, or at least according to [this thing I Googled](http://vrilology.org/FUTHARK.htm).

The symbology intended is that dispatcher is a 'bridge' between the scheduling goroutine and the worker goroutine.

Honestly I just needed a cool Norse thing to name this, I was reaching.  Not to be taken too seriously.

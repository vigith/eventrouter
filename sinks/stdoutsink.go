/*
Copyright 2017 Heptio Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package sinks

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"os"
	"sync"

	"k8s.io/api/core/v1"
)

// let's have few go routines to emit out to stdout.
const stdoutSinkConcurrency = 5

// StdoutSink is the other basic sink
// By default, Fluentd/ElasticSearch won't index glog formatted lines
// By logging raw JSON to stdout, we will get automated indexing which
// can be queried in Kibana.
type StdoutSink struct {
	updateChan chan UpdateEvent
	namespace  string
}

// NewStdoutSink will create a new StdoutSink with default options, returned as
// an EventSinkInterface
func NewStdoutSink(ctx context.Context, namespace string) EventSinkInterface {
	gs := &StdoutSink{
		namespace:  namespace,
		updateChan: make(chan UpdateEvent),
	}

	var wg sync.WaitGroup
	wg.Add(stdoutSinkConcurrency)
	glog.V(3).Infof("Starting stdout sink with concurrency=%d", stdoutSinkConcurrency)

	// let's have couple of parallel routines
	go func() {
		defer wg.Done()
		gs.updateEvents(ctx)
	}()

	// wait
	go func() {
		wg.Wait()
		glog.V(3).Info("Stopping stdout sink WaitGroup")
		close(gs.updateChan)
	}()

	return gs
}

// UpdateEvents implements the EventSinkInterface for stdout. Same like glog sink, this is not a non-blocking call because the
// channel could get full. But ATM I do not care because stdout just logs the message to stdout. It is CPU heavy (JSON Marshalling)
// and has no I/O. So the time complexity of the blocking call is very minimal. Also we could spawn more routines of
// updateEvents to make it concurrent.
func (ss *StdoutSink) UpdateEvents(eNew *v1.Event, eOld *v1.Event) {
	ss.updateChan <- UpdateEvent{
		eNew: eNew,
		eOld: eOld,
	}
}

// UpdateEvents implements the EventSinkInterface
func (ss *StdoutSink) updateEvents(ctx context.Context) {
	for {
		select {
		case event := <-ss.updateChan:
			eData := NewEventData(event.eNew, event.eOld)
			if len(ss.namespace) > 0 {
				namespacedData := map[string]interface{}{}
				namespacedData[ss.namespace] = eData
				if eJSONBytes, err := json.Marshal(namespacedData); err == nil {
					fmt.Println(string(eJSONBytes))
				} else {
					// no point handing err, we have a bigger problem, should we panic?
					_, _ = fmt.Fprintf(os.Stderr, "Failed to json serialize event: %v\n", err)
				}
			} else {
				if eJSONBytes, err := json.Marshal(eData); err == nil {
					fmt.Println(string(eJSONBytes))
				} else {
					// no point handing err, we have a bigger problem, should we panic?
					_, _ = fmt.Fprintf(os.Stderr, "Failed to json serialize event: %v\n", err)
				}
			}
		case <-ctx.Done():
			// no point handing err, we have a bigger problem, should we panic?
			glog.V(3).Infof("Ending, stdout sink receiver channel, got ctx.Done()\n")
			return
		}
	}
}

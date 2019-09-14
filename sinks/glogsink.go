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
	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	"sync"
)

// let's have few go routines to emit out glog.
const glogSinkConcurrency = 5

// GlogSink is the most basic sink
// Useful when you already have ELK/EFK Stack
type GlogSink struct {
	updateChan chan UpdateEvent
}

// NewGlogSink will create a new
func NewGlogSink(ctx context.Context) EventSinkInterface {
	gs := &GlogSink{
		updateChan: make(chan UpdateEvent),
	}

	var wg sync.WaitGroup
	wg.Add(glogSinkConcurrency)
	glog.V(3).Infof("Starting glog sink with concurrency=%d", glogSinkConcurrency)
	// let's have couple of parallel routines
	go func() {
		defer wg.Done()
		gs.updateEvents(ctx)
	}()

	// wait
	go func() {
		wg.Wait()
		glog.V(3).Info("Stopping glog sink WaitGroup")
		close(gs.updateChan)
	}()

	return gs
}

// UpdateEvents implements the EventSinkInterface.
// This is not a non-blocking call because the channel could get full. But ATM I do not care because
// glog just logs the message. It is CPU heavy (JSON Marshalling) and has no I/O. So the time complexity of the
// blocking call is very minimal. Also we could spawn more routines of updateEvents to make it concurrent.
func (gs *GlogSink) UpdateEvents(eNew *v1.Event, eOld *v1.Event) {
	gs.updateChan <- UpdateEvent{
		eNew: eNew,
		eOld: eOld,
	}
}

func (gs *GlogSink) updateEvents(ctx context.Context) {
	for {
		select {
		case event := <-gs.updateChan:
			eData := NewEventData(event.eNew, event.eOld)
			if eJSONBytes, err := json.Marshal(eData); err == nil {
				glog.Info(string(eJSONBytes))
			} else {
				glog.Warningf("Failed to json serialize event: %v", err)
			}
		case <-ctx.Done():
			glog.Warning("Ending, glog sink receiver channel, got ctx.Done()")
			return
		}
	}
}

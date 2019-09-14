package sinks

import (
	"bytes"
	"context"
	"io"
	"k8s.io/client-go/kubernetes/fake"
	"os"
	"testing"
	"time"
)

func TestNewStdoutSink(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var err error

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Errorf("os.Pipe Failed, %s", err)
	}
	os.Stdout = w

	ss := NewStdoutSink(ctx, "")

	client := fake.NewSimpleClientset()

	// test create
	e := createTestEvent(t, client)
	ss.UpdateEvents(e, nil)

	// test update
	oldE, newE := updateTestEvent(t, client, e)
	ss.UpdateEvents(newE, oldE)

	// we are not using it
	deleteTestEvent(t, client)

	cancel()

	err = w.Close()
	if err != nil {
		t.Errorf("Close Failed, %s", err)
	}

	var buf bytes.Buffer
	var n int64
	n, err = io.Copy(&buf, r)
	if err != nil {
		t.Errorf("io.Copy Failed (written=%d), %s", n, err)
	}

	os.Stdout = old

	expected := `{"verb":"ADDED","event":{"metadata":{"name":"my-test-event","namespace":"test-ns","creationTimestamp":null},"involvedObject":{},"reason":"foobar","source":{},"firstTimestamp":null,"lastTimestamp":null,"eventTime":null,"reportingComponent":"","reportingInstance":""}}
{"verb":"UPDATED","event":{"metadata":{"name":"my-test-event","namespace":"test-ns","creationTimestamp":null},"involvedObject":{},"reason":"foobar","source":{},"firstTimestamp":null,"lastTimestamp":null,"eventTime":null,"reportingComponent":"","reportingInstance":""},"old_event":{"metadata":{"name":"my-test-event","namespace":"test-ns","creationTimestamp":null},"involvedObject":{},"reason":"foobar","source":{},"firstTimestamp":null,"lastTimestamp":null,"eventTime":null,"reportingComponent":"","reportingInstance":""}}
`
	// will take a while to flush?
	// retry couple of times to make sure it get's flushed
	var flag bool
	var count = 3
	for count > 0 {
		count--
		if expected == buf.String() {
			flag = true
			break
		}
		// not sure how to do it right? this whole stdout testing!
		time.Sleep(10 * time.Millisecond)
	}

	if !flag {
		t.Errorf("Expected=%s\nGot=%s", expected, buf.String())
	}

}

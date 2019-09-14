package sinks

import (
	"context"
	core_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
)

func TestNewGlogSink(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	gs := NewGlogSink(ctx)

	client := fake.NewSimpleClientset()

	// test create
	e := createTestEvent(t, client)
	gs.UpdateEvents(e, nil)

	// test update
	oldE, newE := updateTestEvent(t, client, e)
	gs.UpdateEvents(newE, oldE)

	// we are not using it
	deleteTestEvent(t, client)

	cancel()
}

func createTestEvent(t *testing.T, client *fake.Clientset) *core_v1.Event {
	var err error
	e := &core_v1.Event{ObjectMeta: meta_v1.ObjectMeta{Name: "my-test-event"}}
	e, err = client.CoreV1().Events("test-ns").Create(e)
	if err != nil {
		t.Errorf("Create Event Failed, %s", err)
	}

	return e
}

func updateTestEvent(t *testing.T, client *fake.Clientset, e *core_v1.Event) (*core_v1.Event, *core_v1.Event) {
	var err error

	// send a update event
	e.Reason = "foobar"
	newE, err := client.CoreV1().Events("test-ns").Update(e)
	if err != nil {
		t.Errorf("Update Event Failed, %s", err)
	}

	return e, newE
}

func deleteTestEvent(t *testing.T, client *fake.Clientset) {
	var err error

	err = client.CoreV1().Events("test-ns").Delete("my-test-event", &meta_v1.DeleteOptions{})
	if err != nil {
		t.Errorf("Delete Event Failed, %s", err)
	}

}

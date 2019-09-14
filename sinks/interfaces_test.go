package sinks

import (
	"bytes"
	"context"
	"fmt"
	"github.com/spf13/viper"
	"reflect"
	"testing"
)

// minimum tests for what I am changing. Please add more as we edit more sinks

func Test_v1ManufactureSink(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	type args struct {
		s string
	}
	tests := []struct {
		name  string
		args  args
		wantE EventSinkInterface
	}{
		{
			name:  "glog",
			args:  args{s: "glog"},
			wantE: NewGlogSink(ctx),
		},
		{
			name:  "stdout",
			args:  args{s: "stdout"},
			wantE: NewStdoutSink(ctx, ""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotE := v1ManufactureSink(ctx, tt.args.s)
			// let's check for the type
			if !reflect.DeepEqual(fmt.Sprintf("%T", gotE), fmt.Sprintf("%T", tt.wantE)) {
				t.Errorf("oldManufactureSink() = %v, want %v", gotE, tt.wantE)
			}
		})
	}
	cancel()
}

func Test_v2ManufactureSink(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	type args struct {
		sinks []string
	}
	tests := []struct {
		name  string
		args  args
		wantE []EventSinkInterface
	}{
		{
			name:  "golog",
			args:  args{sinks: []string{"glog"}},
			wantE: []EventSinkInterface{NewGlogSink(ctx)},
		},
		{
			name:  "stdout",
			args:  args{sinks: []string{"stdout"}},
			wantE: []EventSinkInterface{NewStdoutSink(ctx, "")},
		},
		{
			name:  "glog_stdout",
			args:  args{sinks: []string{"glog", "stdout"}},
			wantE: []EventSinkInterface{NewGlogSink(ctx), NewStdoutSink(ctx, "")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotE := v2ManufactureSink(ctx, tt.args.sinks)
			if len(gotE) != len(tt.wantE) {
				t.Errorf("v2ManufactureSink() len = %d, want len %d", len(gotE), len(tt.wantE))
			} else {
				for i := range gotE {
					if !reflect.DeepEqual(fmt.Sprintf("%T", gotE[i]), fmt.Sprintf("%T", tt.wantE[i])) {
						t.Errorf("v2ManufactureSink() Type = %T, want len %T", gotE[i], tt.wantE[i])
					}
				}
			}

		})
	}

	cancel()
}

func Test_v2ManufactureSink_JSON(t *testing.T) {
	var json = []byte(`
{
  "kubeconfig": "/var/run/kubernetes/admin.kubeconfig",
  "sink": "glog"
}
`)

	viper.SetConfigType("json")
	err := viper.ReadConfig(bytes.NewBuffer(json))
	if err != nil {
		t.Fatalf("viper.ReadConfig Failed, %s", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	gotE := ManufactureSink(ctx)
	wantE := []EventSinkInterface{NewGlogSink(ctx)}
	if len(gotE) != len(wantE) {
		t.Errorf("v2ManufactureSink() len = %d, want len %d", len(gotE), len(wantE))
	} else {
		for i := range gotE {
			if !reflect.DeepEqual(fmt.Sprintf("%T", gotE[i]), fmt.Sprintf("%T", wantE[i])) {
				t.Errorf("v2ManufactureSink() Type = %T, want len %T", gotE[i], wantE[i])
			}
		}
	}

	cancel()
}

func Test_v2ManufactureSink_NestedJSON(t *testing.T) {
	var json = []byte(`
{
  "kubeconfig": "/var/run/kubernetes/admin.kubeconfig",
  "sinks": ["glog", "stdout"],
  "stdout" : {
    "JSONNamespace": "foobar"
  }
}
`)
	viper.SetConfigType("json")
	err := viper.ReadConfig(bytes.NewBuffer(json))
	if err != nil {
		t.Fatalf("viper.ReadConfig Failed, %s", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	gotE := ManufactureSink(ctx)
	// this is not a valid test, the argument is not really tested
	wantE := []EventSinkInterface{NewGlogSink(ctx), NewStdoutSink(ctx, "foobar")}
	if len(gotE) != len(wantE) {
		t.Errorf("v2ManufactureSink() len = %d, want len %d", len(gotE), len(wantE))
	} else {
		for i := range gotE {
			if !reflect.DeepEqual(fmt.Sprintf("%T", gotE[i]), fmt.Sprintf("%T", wantE[i])) {
				t.Errorf("v2ManufactureSink() Type = %T, want len %T", gotE[i], wantE[i])
			}
		}
	}

	cancel()
}

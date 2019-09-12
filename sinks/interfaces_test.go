package sinks

import (
	"bytes"
	"github.com/spf13/viper"
	"reflect"
	"testing"
)

func Test_oldManufactureSink(t *testing.T) {
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
			wantE: NewGlogSink(),
		},
		{
			name:  "stdout",
			args:  args{s: "stdout"},
			wantE: NewStdoutSink(""),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotE := v1ManufactureSink(tt.args.s); !reflect.DeepEqual(gotE, tt.wantE) {
				t.Errorf("oldManufactureSink() = %v, want %v", gotE, tt.wantE)
			}
		})
	}
}

func Test_v2ManufactureSink(t *testing.T) {
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
			wantE: []EventSinkInterface{NewGlogSink()},
		},
		{
			name:  "stdout",
			args:  args{sinks: []string{"stdout"}},
			wantE: []EventSinkInterface{NewStdoutSink("")},
		},
		{
			name:  "glog_stdout",
			args:  args{sinks: []string{"glog", "stdout"}},
			wantE: []EventSinkInterface{NewGlogSink(), NewStdoutSink("")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotE := v2ManufactureSink(tt.args.sinks); !reflect.DeepEqual(gotE, tt.wantE) {
				t.Errorf("v2ManufactureSink() = %v, want %v", gotE, tt.wantE)
			}
		})
	}
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

	got := ManufactureSink()
	wanted := []EventSinkInterface{NewGlogSink()}
	if !reflect.DeepEqual(got, wanted) {
		t.Errorf("v2ManufactureSink() = %v, want %v", got, wanted)
	}
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

	got := ManufactureSink()
	wanted := []EventSinkInterface{NewGlogSink(), NewStdoutSink("foobar")}
	if !reflect.DeepEqual(got, wanted) {
		t.Errorf("v2ManufactureSink() = %v, want %v", got, wanted)
	}

}

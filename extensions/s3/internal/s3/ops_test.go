// v0.7.1
// Author: wunderbarb
// Sep 2025

package s3

import (
	"context"
	"testing"
	"time"

	"github.com/wunderbarb/test"
)

func Test_Join(t *testing.T) {
	tests := []struct {
		els    []string
		result string
	}{
		{[]string{}, ""},
		{[]string{"bkt", "prefix", "object"}, "s3://bkt/prefix/object"},
		{[]string{"s3://bkt", "prefix", "object"}, "s3://bkt/prefix/object"},
		{[]string{"s3://bkt", "prefix", "", "object"}, "s3://bkt/prefix/object"},
		{[]string{"s3://bkt", "prefix", "..", "object"}, "s3://bkt/object"},
	}
	for i, tt := range tests {
		if tt.result != Join(tt.els...) {
			t.Errorf("expected %s got %v sample %d", tt.result, tt.result, i+1)
		}
	}
}

func Test_Size(t *testing.T) {
	s, err := Size(context.Background(), _cGolden100K)
	if err != nil {
		t.Fatal(err)
	}
	if s != int64(100*K) {
		t.Errorf("expected %v got %v", int64(100*K), s)
	}
	_, err = Size(context.Background(), _cs3Testdata+"/bad")
	if err == nil {
		t.Error("expected error")
	}
}

func TestHasAccess(t *testing.T) {
	if HasAccess(context.Background(), _cS3) == false {
		t.Fatal("expected true")
	}
	if HasAccess(context.Background(), Join(_cS3, test.RandomID()),
		WithRetry(3, 500*time.Millisecond)) == false {
		t.Fatal("expected true")
	}
	if HasAccess(context.Background(), "s3://bad") == true {
		t.Fatal("expected false")
	}
	if HasAccess(context.Background(), "s3://bad/bad", WithRetry(3, 500*time.Millisecond)) == true {
		t.Fatal("expected false")
	}
}

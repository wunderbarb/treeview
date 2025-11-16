package s3

import (
	"context"
	"testing"
)

func Test_IsDir(t *testing.T) {
	tests := []struct {
		path       string
		expSuccess bool
	}{
		{_cs3Testdata + "/golden", true},
		{_cs3Testdata + "/golden/", true},
		{_cs3Testdata + "/golden/sample", false},
		{_cs3Testdata, true},
		{"s3://bad", false},
	}
	for i, tt := range tests {
		if tt.expSuccess != IsDir(context.Background(), tt.path) {
			t.Errorf("%d: expected %v, got %v", i, tt.expSuccess, IsDir(context.Background(), tt.path))
		}
	}
}

func Test_Base(t *testing.T) {
	tests := []struct {
		path   string
		result string
	}{
		{"s3://bucket/prefix/object/", "object"},
		{"s3://bucket/prefix", "prefix"},
		{"s3://bucket", ""},
		{"", "."},
	}
	for _, tt := range tests {
		if tt.result != Base(tt.path) {
			t.Errorf("%s: expected %s, got %s", tt.path, tt.result, tt.result)
		}
	}
}

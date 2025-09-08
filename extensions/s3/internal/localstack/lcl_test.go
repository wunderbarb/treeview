// v0.1.1
// Author: wunderbarb
// Sep 2025

package localstack

import (
	"testing"
)

func TestUseNot(t *testing.T) {
	defer func() { _ = Use() }()

	if InUse() == false {
		t.Fatal("InUse() should return true")
	}
	if err := UseNot(); err != nil {
		t.Fatalf("UseNot() should not return error, but got %v", err)
	}
	if InUse() == true {
		t.Error("UseNot() should return false")
	}
	if err := Use(); err != nil {
		t.Fatalf("Use should not return error, but got %v", err)
	}
	if InUse() == false {
		t.Fatal("InUse() should return true")
	}
}

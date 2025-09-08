// v0.1.12
// Author: wunderbarb
// Sep 2025

package localstack

import (
	"math/rand/v2"
	"os"
	"path/filepath"
	"testing"
)

func TestMain(m *testing.M) {
	isPanic(Use())
	defer func() { _ = UseNot() }()
	_ = m.Run()
}

func TestIsRunning(t *testing.T) {
	if IsRunning() == false {
		t.Fatal("expected true")
	}
}

func TestCreateBucket(t *testing.T) {
	bkt := "tst" + randomID()
	if err := CreateBucket(bkt); err != nil {
		t.Fatal(err)
	}
	if err := CreateBucket(bkt); err == nil {
		t.Error(err)
	}
	if err := CreateBucket(bkt, WithNoErrorIfExist()); err != nil {
		t.Fatal(err)
	}
	if err := DeleteBucket(bkt); err != nil {
		t.Error(err)
	}
	if err := DeleteBucket("s3://bad"); err == nil {
		t.Error("should generate an error")
	}
	if err := CreateBucket("tst" + randomID()); err == nil {
		t.Error("should generate an error")
	}
}

func TestCreateBucketAt(t *testing.T) {
	bkt := "tst" + randomID()
	const cRegion = "us-west-2"
	if err := CreateBucketAt(bkt, cRegion); err != nil {
		t.Fatal(err)
	}
	if err := DeleteBucketAt(bkt, cRegion); err != nil {
		t.Error(err)
	}
}

func TestPutObject(t *testing.T) {
	bkt := "tst" + randomID()
	isPanic(CreateBucket(bkt))
	defer func() { _ = DeleteBucket(bkt) }()

	hDir, err := os.UserHomeDir()
	isPanic(err)
	if err = PutObject("s3://"+bkt+"/test", filepath.Join(hDir, "Dev", "golden", "sample100K.golden")); err != nil {
		t.Fatal(err)
	}
	if err = PutObject("s3://"+bkt+"/test", filepath.Join(hDir, "Dev", "golden", "sample100K.golden")); err != nil {
		t.Error(err)
	}
}

func TestDeleteBucket(t *testing.T) {
	bkt := "tst" + randomID()
	isPanic(CreateBucket(bkt))
	hDir, _ := os.UserHomeDir()
	isPanic(PutObject("s3://"+bkt+"/test", filepath.Join(hDir, "Dev", "golden", "sample100K.golden")))

	if err := DeleteBucket(bkt); err != nil {
		t.Error(err)
	}
}

func TestSetEndPointURLForLambda(t *testing.T) {
	const key = "LOCALSTACK_HOSTNAME"

	adr := "127.27.0.2"
	isPanic(os.Setenv(key, adr))
	defer func() {
		_ = os.Unsetenv(key)
		_endpointURL = "--endpoint-url=http://localhost:4566"
	}()

	SetEndPointURLForLambda()
	if "--endpoint-url=http://"+adr+":4566" != _endpointURL {
		t.Error("should be equal")
	}
	if GetEndPointURL() != "http://"+adr+":4566" {
		t.Error("should be equal")
	}
}

func isPanic(err error) {
	if err != nil {
		panic(err)
	}
}

// randomID returns a random 16-character, alphanumeric, ID.
func randomID() string {
	const sizeID = 16
	size := sizeID
	var buffer []byte
	choice := []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz01234567890")
	choiceSize := len(choice)
	for i := 0; i < size; i++ {
		// generates the characters
		s := rand.IntN(choiceSize)
		buffer = append(buffer, choice[s])
	}
	return string(buffer)
}

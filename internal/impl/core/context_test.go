package core

import (
	"clu-vms/internal/spec"
	"testing"
)

func TestWorkContext(t *testing.T) {
	wc, err := NewLocalContext("testdata")
	if err != nil {
		t.Fatal(err)
	}
	if wc.GetType() != spec.WorkContextTypeLocal {
		t.Fatal("wrong work context type")
	}

	err = wc.CreateDir("dummy_dir")
	if err != nil {
		t.Fatal("failed to create directory")
	}

	err = wc.CreateFile("dummy_dir/dummy_file.txt", "Hello, World!")
	if err != nil {
		t.Fatal("failed to create file")
	}

	if !wc.FileExists("dummy_dir/dummy_file.txt") {
		t.Fatal("file does not exist")
	}

	err = wc.DeleteDir("dummy_dir")
	if err != nil {
		t.Fatal("failed to delete directory")
	}
}

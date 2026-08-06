package core

import (
	"clu-vms/internal/spec"
	"testing"
)

func TestOSImages(t *testing.T) {
	workCtx, err := NewLocalContext("testdata")
	if err != nil {
		t.Fatal(err)
	}
	osImgs := NewOSImages(workCtx)

	err = osImgs.DeleteAllImages()
	if err != nil {
		t.Fatal(err)
	}

	osImgExists := osImgs.OSExists(spec.Alpine_3_24)
	if osImgExists {
		t.Fatal("image exists")
	}

	downloaded, err := osImgs.DownloadLatest(spec.Alpine_3_24)
	if err != nil {
		t.Fatal(err)
	}
	if !downloaded {
		t.Fatal("image not downloaded")
	}

	err = osImgs.CopyOSImage(spec.Alpine_3_24, "my-alpine.qcow2")
	if err != nil {
		t.Fatal(err)
	}

	if !workCtx.FileExists("my-alpine.qcow2") {
		t.Fatal("file does not exist")
	}
}

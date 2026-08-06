package core

import (
	"clu-vms/internal/spec"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

type osImage struct {
	name     string
	filepath string
	url      string
	osType   spec.OSImagesType
}

var osImages map[spec.OSImagesType]osImage

const imgsPath = ".assets/imgs/"

func newOsImage(osType spec.OSImagesType, url string) osImage {
	osName, _, _ := strings.Cut(string(osType), "_")
	return osImage{
		filepath: filepath.Join(imgsPath, osName+".qcow2"),
		url:      url,
		name:     string(osType),
		osType:   osType,
	}
}

func init() {
	osImages = make(map[spec.OSImagesType]osImage)
	osImages[spec.Alpine_3_24] = newOsImage(spec.Alpine_3_24, "https://dl-cdn.alpinelinux.org/alpine/v3.24/releases/cloud/generic_alpine-3.24.1-x86_64-bios-cloudinit-r0.qcow2")
}

type OSImages struct {
	workCtx spec.WorkContext
}

func NewOSImages(workCtx spec.WorkContext) *OSImages {
	return &OSImages{
		workCtx: workCtx,
	}
}

func (oi *OSImages) DownloadLatest(osImage spec.OSImagesType) (bool, error) {
	osImg, ok := osImages[osImage]
	if !ok {
		return false, errors.New("OS image not found")
	}

	if oi.workCtx.FileExists(osImg.filepath) {
		return false, nil
	}

	return true, oi.workCtx.DownloadFile(osImg.url, osImg.filepath)
}

func (oi *OSImages) OSExists(osImage spec.OSImagesType) bool {
	osImg, ok := osImages[osImage]
	if !ok {
		return false
	}

	return oi.workCtx.FileExists(osImg.filepath)
}

func (oi *OSImages) CopyOSImage(osImage spec.OSImagesType, targetPath string) error {
	osImg, ok := osImages[osImage]
	if !ok {
		return errors.New("OS image not found")
	}

	if !oi.workCtx.FileExists(osImg.filepath) {
		return errors.New("OS image does not exist")
	}

	if oi.workCtx.FileExists(targetPath) {
		err := oi.workCtx.DeleteFile(targetPath)
		if err != nil {
			return fmt.Errorf("failed to delete file: %w", err)
		}
	}

	err := oi.workCtx.CreateDir(filepath.Dir(targetPath))
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return oi.workCtx.CopyFile(osImg.filepath, targetPath)
}

func (oi *OSImages) DeleteImage(osImage spec.OSImagesType) error {
	osImg, ok := osImages[osImage]
	if !ok {
		return errors.New("OS image not found")
	}

	return oi.workCtx.DeleteFile(osImg.filepath)
}

func (oi *OSImages) DeleteAllImages() error {
	return oi.workCtx.DeleteDir(imgsPath)
}

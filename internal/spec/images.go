package spec

type OSImagesFamily string

const (
	Alpine = OSImagesFamily("alpine")
)

type OSImagesType struct {
	OS      OSImagesFamily
	Version string
}

var (
	alpine_3_24 = OSImagesType{
		OS:      Alpine,
		Version: "3.24",
	}
)

func Alpine_3_24() OSImagesType {
	return alpine_3_24
}

type OSImages interface {
	DeleteImage(osImage OSImagesType) error
	DeleteAllImages() error
	DownloadLatest(osImage OSImagesType) (bool, error)
	OSExists(osImage OSImagesType) bool
	CopyOSImage(osImage OSImagesType, targetPath string) error
}

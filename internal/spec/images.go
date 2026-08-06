package spec

type OSImagesType string

const (
	Alpine_3_24 = OSImagesType("alpine_3_24")
)

type OSImages interface {
	DeleteImage(osImage OSImagesType) error
	DeleteAllImages() error
	DownloadLatest(osImage OSImagesType) (bool, error)
	OSExists(osImage OSImagesType) bool
	CopyOSImage(osImage OSImagesType, targetPath string) error
}

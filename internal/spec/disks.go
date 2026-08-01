package spec

type FileSystem int

const (
	FileSystemExt4 FileSystem = iota
	FileSystemXFS
)

type DiskService interface {
	CreateDataDisk(path string, sizeGB int, fs FileSystem) error
	DataDiskExists(path string) bool
	DeleteDataDisk(path string) error
}

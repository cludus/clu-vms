package spec

type WorkContextType int

const (
	WorkContextTypeLocal WorkContextType = iota
	WorkContextTypeRemote
)

type WorkContextRemoteInfo struct {
	Addr    string
	User    string
	SSHKey  string
	SSHPort int
}

type WorkContext interface {
	GetType() WorkContextType
	RunCommand(cmd ...string) error
	RunCommandWithOutput(cmd ...string) (string, error)
	WorkDir() string
	CreateDir(path string) error
	DeleteDir(path string) error
	CreateFile(path string, content string) error
	RemoteInfo() WorkContextRemoteInfo
	DownloadFile(url string, path string) error
	FileExists(path string) bool
	DirExists(path string) bool
	CopyFile(source string, target string) error
	DeleteFile(path string) error
	ReadFile(path string) (string, error)
	WriteFile(path string, content string) error
}

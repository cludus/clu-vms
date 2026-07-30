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
	RunCommand(cmd []string) error
	WorkDir() string
	CreateDir(path string)
	CreateFile(path string, content string)
	RemoteInfo() WorkContextRemoteInfo
}

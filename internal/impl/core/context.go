package core

import (
	"clu-vms/internal/spec"
	"fmt"
	"os/exec"
)

type WorkContext struct {
	workDir string
	wcType  spec.WorkContextType
}

func NewLocalContext(workDir string) *WorkContext {
	return &WorkContext{
		workDir: workDir,
		wcType:  spec.WorkContextTypeLocal,
	}
}

func (wc *WorkContext) GetType() spec.WorkContextType {
	return wc.wcType
}

func (wc *WorkContext) RunCommand(cmdStr []string) error {
	cmd := exec.Command(cmdStr[0], cmdStr[1:]...)

	// Runs the command and captures combined stdout/stderr
	output, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}

func (wc *WorkContext) WorkDir() string {
	return wc.workDir
}

func (wc *WorkContext) CreateDir(path string) {

}

func (wc *WorkContext) CreateFile(path string, content string) {

}

func (wc *WorkContext) RemoteInfo() spec.WorkContextRemoteInfo {
	return spec.WorkContextRemoteInfo{}
}

package core

import (
	"clu-vms/internal/spec"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

type WorkContextImpl struct {
	workDir    string
	wcType     spec.WorkContextType
	remoteInfo spec.WorkContextRemoteInfo
}

func NewLocalContext(workDir string) (*WorkContextImpl, error) {
	absFP, err := filepath.Abs(workDir)
	if err != nil {
		return nil, err
	}
	result := WorkContextImpl{
		workDir: absFP,
		wcType:  spec.WorkContextTypeLocal,
	}
	return &result, nil
}

func (wc *WorkContextImpl) GetType() spec.WorkContextType {
	return wc.wcType
}

func (wc *WorkContextImpl) RunCommandWithOutput(cmdStr ...string) (string, error) {
	cmd := exec.Command(cmdStr[0], cmdStr[1:]...)
	cmd.Dir = wc.workDir
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func (wc *WorkContextImpl) RunCommand(cmdStr ...string) error {
	cmd := exec.Command(cmdStr[0], cmdStr[1:]...)
	cmd.Dir = wc.workDir

	// Runs the command and captures combined stdout/stderr
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(output))
		return err
	}
	fmt.Println(string(output))
	return nil
}

func (wc *WorkContextImpl) WorkDir() string {
	return wc.workDir
}

func (wc *WorkContextImpl) CreateDir(path string) error {
	return os.MkdirAll(filepath.Join(wc.workDir, path), 0755)
}

func (wc *WorkContextImpl) DeleteDir(path string) error {
	// Deletes the directory and everything inside it
	return os.RemoveAll(filepath.Join(wc.workDir, path))
}

func (wc *WorkContextImpl) CreateFile(path string, content string) error {
	file, err := os.Create(filepath.Join(wc.workDir, path))
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(content)
	return err
}

func (wc *WorkContextImpl) RemoteInfo() spec.WorkContextRemoteInfo {
	return wc.remoteInfo
}

func (wc *WorkContextImpl) DownloadFile(url string, target string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	dir := filepath.Dir(target)
	if !wc.DirExists(dir) {
		err = wc.CreateDir(dir)
		if err != nil {
			return err
		}
	}

	out, err := os.Create(filepath.Join(wc.workDir, target))
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func (wc *WorkContextImpl) FileExists(path string) bool {
	info, err := os.Stat(filepath.Join(wc.workDir, path))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false
		}
		fmt.Printf("An error occurred: %v\n", err)
		return false
	}

	return !info.IsDir()
}

func (wc *WorkContextImpl) CopyFile(src string, dst string) error {
	// Open the source file for reading
	source, err := os.Open(filepath.Join(wc.workDir, src))
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer source.Close()

	// Create the destination file (truncates if it already exists)
	destination, err := os.Create(filepath.Join(wc.workDir, dst))
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destination.Close()

	// Stream the data from source to destination
	_, err = io.Copy(destination, source)
	if err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}

	// Commit the file contents to stable storage
	err = destination.Sync()
	if err != nil {
		return fmt.Errorf("failed to sync destination file: %w", err)
	}

	return nil
}

func (wc *WorkContextImpl) DeleteFile(path string) error {
	return os.Remove(filepath.Join(wc.workDir, path))
}

func (wc *WorkContextImpl) DirExists(path string) bool {
	info, err := os.Stat(filepath.Join(wc.workDir, path))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false
		}
		fmt.Printf("An error occurred: %v\n", err)
		return false
	}

	return info.IsDir()
}

func (wc *WorkContextImpl) ReadFile(path string) (string, error) {
	bytes, err := os.ReadFile(filepath.Join(wc.workDir, path))
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (wc *WorkContextImpl) WriteFile(path string, content string) error {
	return os.WriteFile(filepath.Join(wc.workDir, path), []byte(content), os.FileMode.Perm(0644))
}

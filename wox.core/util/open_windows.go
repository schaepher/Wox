package util

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func ShellOpen(path string) error {
	return exec.Command("cmd", "/C", "start", "explorer.exe", path).Start()
}

func ShellRun(name string, arg ...string) (*exec.Cmd, error) {
	cmd := exec.Command(name, arg...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} // Hide the window
	cmd.Stdout = GetLogger().GetWriter()
	cmd.Stderr = GetLogger().GetWriter()
	cmd.Env = append(os.Environ(), "PYTHONIOENCODING=utf-8")
	cmdErr := cmd.Start()
	if cmdErr != nil {
		return nil, cmdErr
	}

	return cmd, nil
}

func ShellRunWithEnv(name string, envs []string, arg ...string) (*exec.Cmd, error) {
	if len(envs) == 0 {
		return ShellRun(name, arg...)
	}

	cmd := exec.Command(name, arg...)
	cmd.Stdout = GetLogger().GetWriter()
	cmd.Stderr = GetLogger().GetWriter()
	cmd.Env = append(os.Environ(), envs...)
	cmdErr := cmd.Start()
	if cmdErr != nil {
		return nil, cmdErr
	}

	return cmd, nil
}

func ShellRunOutput(name string, arg ...string) ([]byte, error) {
	cmd := exec.Command(name, arg...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} // Hide the window
	return cmd.Output()
}

func ShellOpenFileInFolder(path string) error {
	if _, err := os.Stat(path); err != nil {
		return err
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	return exec.Command("explorer.exe", "/select,", absPath).Start()
}

func OpenHttp(path string) error {
	return exec.Command("cmd", "/C", "start", path).Start()
}
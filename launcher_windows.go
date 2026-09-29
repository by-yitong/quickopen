//go:build windows

package main

import (
	"os/exec"
	"syscall"
)

// startProcess 拉起进程;Windows 下隐藏控制台窗口。
func startProcess(argv []string) error {
	if len(argv) == 0 {
		return nil
	}
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Start()
}

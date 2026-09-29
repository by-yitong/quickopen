//go:build !windows

package main

import (
	"os/exec"
	"syscall"
)

// startProcess 拉起进程;Setsid 让子进程脱离父进程,窗口关闭后仍存活。
func startProcess(argv []string) error {
	if len(argv) == 0 {
		return nil
	}
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		return err
	}
	// 异步回收子进程,避免每次打开都留下一个僵尸进程
	go func() { _ = cmd.Wait() }()
	return nil
}

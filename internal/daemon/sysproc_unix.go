//go:build !windows

package daemon

import "syscall"

// sysProcAttr 让子进程成为新会话的首进程（setsid），从而脱离控制终端：
// 终端关闭时不会再收到 SIGHUP，也不会被终端的作业控制影响。
func sysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setsid: true}
}

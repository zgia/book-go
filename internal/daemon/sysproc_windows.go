//go:build windows

package daemon

import "syscall"

// sysProcAttr Windows 上没有 setsid，子进程本身不会随父进程退出，
// 这里不做额外处理。
func sysProcAttr() *syscall.SysProcAttr {
	return nil
}

// Package daemon 实现进程自后台化（self-daemonize）。
//
// 用法：在 main 的最开始调用 Requested 判断是否需要后台运行，需要则调用
// Daemonize 启动子进程并退出父进程；子进程通过环境变量 BOOK_DAEMONIZED
// 识别出来，直接以前台方式继续执行。
package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// envKey 标记当前进程是 Daemonize 拉起的子进程，避免无限递归。
const envKey = "BOOK_DAEMONIZED"

// logFileName 后台进程的标准输出文件，与 nohup 的默认文件名保持一致。
const logFileName = "nohup.out"

// IsChild 报告当前进程是否由 Daemonize 在后台拉起。
func IsChild() bool {
	return os.Getenv(envKey) == "1"
}

// Requested 报告命令行参数是否要求以守护进程方式运行。
// 只对 api 子命令生效：需要同时出现 api 与 --daemon（或 -d）。
func Requested(args []string) bool {
	var hasApi, hasDaemon bool
	for _, arg := range args {
		switch arg {
		case "api":
			hasApi = true
		case "--daemon", "-d", "--daemon=true", "-d=true":
			hasDaemon = true
		}
	}
	return hasApi && hasDaemon
}

// Daemonize 重新拉起一份自身进程，使其脱离当前终端在后台运行。
//
// 子进程会继承父进程的工作目录与命令行参数（含 --daemon），并带上
// BOOK_DAEMONIZED=1；其标准输入重定向到 /dev/null，标准输出与标准错误
// 追加写入当前目录下的 nohup.out。
//
// 调用方在返回 nil 后应立即 os.Exit(0) 结束父进程。
func Daemonize() error {
	exe, err := os.Executable()
	if err != nil || exe == "" {
		exe = os.Args[0]
	}

	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("打开日志文件 %s 失败: %w", logFileName, err)
	}
	defer logFile.Close()

	devNull, err := os.Open(os.DevNull)
	if err != nil {
		return fmt.Errorf("打开 %s 失败: %w", os.DevNull, err)
	}
	defer devNull.Close()

	cmd := exec.Command(exe, os.Args[1:]...)
	cmd.Env = append(os.Environ(), envKey+"=1")
	cmd.Stdin = devNull
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.SysProcAttr = sysProcAttr()

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动后台进程失败: %w", err)
	}

	fmt.Printf("%s 已在后台启动，PID: %d，输出写入 %s\n",
		filepath.Base(exe), cmd.Process.Pid, logFileName)
	return nil
}

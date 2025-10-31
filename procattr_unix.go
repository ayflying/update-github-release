//go:build !windows
// +build !windows

package update

import (
    "os/exec"
    "syscall"
)

func setCmdSysProcAttr(cmd *exec.Cmd) {
    cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}
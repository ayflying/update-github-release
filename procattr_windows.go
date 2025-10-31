//go:build windows
// +build windows

package update

import (
    "os/exec"
)

func setCmdSysProcAttr(cmd *exec.Cmd) {
    // no-op on Windows
}
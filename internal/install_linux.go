//go:build linux
// +build linux

package internal

import (
	"fmt"
	"os"
	"syscall"
	"time"
)

func getInstallTime() (time.Time, error) {
	fi, err := os.Stat("/etc")
	if err != nil {
		return time.Time{}, err
	}
	st, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		return time.Time{}, fmt.Errorf("file sys: %T", fi.Sys())
	}
	return time.Unix(st.Ctim.Sec, st.Ctim.Nsec), nil
}
package sshserver

import (
	"os"
	"syscall"
	"unsafe"
)

type PTYManager struct{}

func NewPTYManager() PTYManager {
	return PTYManager{}
}

func (PTYManager) SetWindowSize(f *os.File, width, height int) {
	syscall.Syscall(
		syscall.SYS_IOCTL,
		f.Fd(),
		uintptr(syscall.TIOCSWINSZ),
		uintptr(unsafe.Pointer(&struct {
			h, w, x, y uint16
		}{
			uint16(height), uint16(width), 0, 0,
		})),
	)
}

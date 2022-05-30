package sshserver

import (
	"os/exec"
	"os/user"
)

type UserCommandManager struct {
	user *user.User
}

func NewUserCommandManager(
	user *user.User,
) UserCommandManager {

	return UserCommandManager{
		user: user,
	}
}

func (u UserCommandManager) Build(args ...string) *exec.Cmd {
	cmdToBuildArgs := []string{
		"--set-home",
		"--login",
		"--user",
		u.user.Username,
	}

	return exec.Command("sudo", append(cmdToBuildArgs, args...)...)
}

func (u UserCommandManager) BuildShell() *exec.Cmd {
	return u.Build()
}

func (u UserCommandManager) BuildShellPTY() *exec.Cmd {
	cmdToBuildArgs := []string{
		"login",
		"-f",
		u.user.Username,
	}

	return exec.Command("sudo", cmdToBuildArgs...)
}

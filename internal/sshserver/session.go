package sshserver

import (
	"log"

	"github.com/gliderlabs/ssh"
	"github.com/recode-sh/agent/constants"
	"github.com/recode-sh/agent/internal/docker"
)

type SessionExecShellManager interface {
	ManageShellInDevEnv(sshSession ssh.Session) error
	ManageShellPTYInDevEnv(sshSession ssh.Session) error
	ManageExecInDevEnv(sshSession ssh.Session) error
	ManageShellPTY(sshSession ssh.Session) error
	ManageShell(sshSession ssh.Session) error
	ManageExec(sshSession ssh.Session) error
}

type Session struct {
	manager SessionExecShellManager
}

func NewSession(
	manager SessionExecShellManager,
) Session {

	return Session{
		manager: manager,
	}
}

func (s Session) Start(sshSession ssh.Session) {
	var sessionError error

	defer func() {
		if sessionError != nil {
			log.Println(sessionError)
			sshSession.Exit(1)
			return
		}

		sshSession.Exit(0)
	}()

	dockerClient, err := docker.NewDefaultClient()

	// We don't handle error here because
	// we want to be able to login to instance via SSH
	// even if the docker container cannot be reached
	if err != nil {
		log.Println(err)
	}

	isContainerRunning, err := docker.IsContainerRunning(
		dockerClient,
		constants.DevEnvDockerContainerName,
	)

	// Same than previous comment
	if err != nil {
		log.Println(err)
	}

	if len(sshSession.Command()) == 0 { // "shell" session
		_, _, hasPTY := sshSession.Pty()

		if hasPTY {
			// if !isContainerRunning {
			// 	sessionError = s.manager.ManageShellPTY(sshSession)
			// 	return
			// }

			sessionError = s.manager.ManageShellPTYInDevEnv(sshSession)
			return
		}

		// if !isContainerRunning {
		// 	sessionError = s.manager.ManageShell(sshSession)
		// 	return
		// }

		sessionError = s.manager.ManageShellInDevEnv(sshSession)
		return
	}

	if !isContainerRunning {
		sessionError = s.manager.ManageExec(sshSession)
		return
	}

	// "exec" session
	sessionError = s.manager.ManageExecInDevEnv(sshSession)
}

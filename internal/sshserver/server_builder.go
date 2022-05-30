package sshserver

import (
	"log"
	"os/user"

	"github.com/gliderlabs/ssh"
)

type ServerAuther interface {
	BuildHostKeySigner() (ssh.Signer, error)
	CheckPublicKeyValidity(username string, passedKey ssh.PublicKey) (bool, error)
}

type ServerBuilder struct {
	auth       ServerAuther
	listenAddr string
}

func NewServerBuilder(
	auth ServerAuther,
	listenAddr string,
) ServerBuilder {

	return ServerBuilder{
		auth:       auth,
		listenAddr: listenAddr,
	}
}

func (s ServerBuilder) Build() (*ssh.Server, error) {
	hostKeySigner, err := s.auth.BuildHostKeySigner()

	if err != nil {
		return nil, err
	}

	forwardHandler := &ssh.ForwardedTCPHandler{}

	return &ssh.Server{
		Addr: s.listenAddr,

		HostSigners: []ssh.Signer{hostKeySigner},

		PublicKeyHandler: func(ctx ssh.Context, passedKey ssh.PublicKey) bool {
			keyIsValid, err := s.auth.CheckPublicKeyValidity(ctx.User(), passedKey)

			if err != nil {
				log.Println(err)
				return false
			}

			return keyIsValid
		},

		Handler: func(sshSession ssh.Session) {
			user, err := user.Lookup(sshSession.User())

			if err != nil {
				log.Println(err)
				sshSession.Close()
				return
			}

			sessionManager := NewSessionManager(
				NewUserCommandManager(user),
				NewPTYManager(),
			)

			session := NewSession(sessionManager)
			session.Start(sshSession)
		},

		LocalPortForwardingCallback: ssh.LocalPortForwardingCallback(func(ctx ssh.Context, dhost string, dport uint32) bool {
			return true
		}),

		ReversePortForwardingCallback: ssh.ReversePortForwardingCallback(func(ctx ssh.Context, host string, port uint32) bool {
			return true
		}),

		RequestHandlers: map[string]ssh.RequestHandler{
			"tcpip-forward":        forwardHandler.HandleSSHRequest,
			"cancel-tcpip-forward": forwardHandler.HandleSSHRequest,
		},

		ChannelHandlers: map[string]ssh.ChannelHandler{
			"direct-tcpip":                   ssh.DirectTCPIPHandler,
			"session":                        ssh.DefaultSessionHandler,
			"direct-streamlocal@openssh.com": handleDirectStreamLocalOpenSSH,
		},
	}, nil
}

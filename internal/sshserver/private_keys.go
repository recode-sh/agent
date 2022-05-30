package sshserver

import (
	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
)

type PrivateKeyManager struct{}

func NewPrivateKeyManager() PrivateKeyManager {
	return PrivateKeyManager{}
}

func (PrivateKeyManager) ParsePrivateKey(
	privateKeyBytes []byte,
) (ssh.Signer, error) {

	return gossh.ParsePrivateKey(privateKeyBytes)
}

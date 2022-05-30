package sshserver

import "github.com/gliderlabs/ssh"

type AuthorizedKeyManager struct{}

func NewAuthorizedKeyManager() AuthorizedKeyManager {
	return AuthorizedKeyManager{}
}

func (AuthorizedKeyManager) ParseAuthorizedKeys(
	authorizedKeysBytes []byte,
) ([]ssh.PublicKey, error) {

	authorizedKeys := []ssh.PublicKey{}

	for len(authorizedKeysBytes) > 0 {
		pubKey, _, _, rest, err := ssh.ParseAuthorizedKey(authorizedKeysBytes)

		if err != nil {
			return nil, err
		}

		authorizedKeys = append(authorizedKeys, pubKey)
		authorizedKeysBytes = rest
	}

	return authorizedKeys, nil
}

func (AuthorizedKeyManager) CheckPublicKeyEqualsAuthorizedKey(
	publicKey ssh.PublicKey,
	authorizedKey ssh.PublicKey,
) bool {

	return ssh.KeysEqual(publicKey, authorizedKey)
}

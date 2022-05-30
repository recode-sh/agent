package sshserver

import (
	"github.com/gliderlabs/ssh"
)

type FileReader interface {
	ReadFile(filePath string) ([]byte, error)
}

type PrivateKeyParser interface {
	ParsePrivateKey(privateKeyBytes []byte) (ssh.Signer, error)
}

type AuthorizedKeyParserChecker interface {
	ParseAuthorizedKeys(authorizedKeysBytes []byte) ([]ssh.PublicKey, error)
	CheckPublicKeyEqualsAuthorizedKey(publicKey, authorizedKey ssh.PublicKey) bool
}

type AuthorizedUser struct {
	UserName               string
	AuthorizedKeysFilePath string
}

type Auth struct {
	fileManager          FileReader
	privateKeyManager    PrivateKeyParser
	authorizedKeyManager AuthorizedKeyParserChecker
	hostKeyFilePath      string
	authorizedUsers      []AuthorizedUser
}

func NewAuth(
	fileManager FileReader,
	privateKeyManager PrivateKeyParser,
	authorizedKeyManager AuthorizedKeyParserChecker,
	hostKeyFilePath string,
	authorizedUsers []AuthorizedUser,
) *Auth {

	return &Auth{
		fileManager:          fileManager,
		privateKeyManager:    privateKeyManager,
		authorizedKeyManager: authorizedKeyManager,
		hostKeyFilePath:      hostKeyFilePath,
		authorizedUsers:      authorizedUsers,
	}
}

func (a *Auth) BuildHostKeySigner() (ssh.Signer, error) {
	hostKey, err := a.fileManager.ReadFile(a.hostKeyFilePath)

	if err != nil {
		return nil, err
	}

	return a.privateKeyManager.ParsePrivateKey(hostKey)
}

func (a *Auth) lookupAuthorizedKeysForUser(username string) ([]ssh.PublicKey, error) {
	for _, authorizedUser := range a.authorizedUsers {
		if authorizedUser.UserName != username {
			continue
		}

		authorizedKeysBytes, err := a.fileManager.ReadFile(
			authorizedUser.AuthorizedKeysFilePath,
		)

		if err != nil {
			return nil, err
		}

		authorizedKeys, err := a.authorizedKeyManager.ParseAuthorizedKeys(
			authorizedKeysBytes,
		)

		if err != nil {
			return nil, err
		}

		return authorizedKeys, nil
	}

	return nil, nil
}

func (a *Auth) CheckPublicKeyValidity(
	username string,
	passedKey ssh.PublicKey,
) (bool, error) {

	authorizedKeys, err := a.lookupAuthorizedKeysForUser(
		username,
	)

	if err != nil {
		return false, err
	}

	if authorizedKeys == nil {
		return false, nil
	}

	for _, authorizedKey := range authorizedKeys {
		if a.authorizedKeyManager.CheckPublicKeyEqualsAuthorizedKey(passedKey, authorizedKey) {
			return true, nil
		}
	}

	return false, nil
}

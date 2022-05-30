package devenv

import (
	"os/user"

	"github.com/recode-sh/agent/constants"
)

func lookupRecodeUserAndDockerGroup() (*user.User, *user.Group, error) {
	recodeUser, err := user.Lookup(constants.DevEnvRecodeUserName)

	if err != nil {
		return nil, nil, err
	}

	dockerGroup, err := user.LookupGroup(constants.DevEnvDockerGroupName)

	if err != nil {
		return nil, nil, err
	}

	return recodeUser, dockerGroup, nil
}

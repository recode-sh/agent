package docker

import (
	"errors"
	"strings"
)

func LookupDockerfileBaseImage(dockerfilePath string) (string, error) {
	dockerfileCmds, err := parseDockerfile(dockerfilePath)

	if err != nil {
		return "", err
	}

	var aliasImageMap = map[string]string{}
	var lastFromCmd dockerfileCmd

	for _, dockerfileCmd := range dockerfileCmds {
		if dockerfileCmd.cmd != "FROM" {
			continue
		}

		lastFromCmd = dockerfileCmd
		fromCmdValue := lastFromCmd.value

		if len(fromCmdValue) != 3 { // FROM "image"
			continue
		}

		// FROM "image" AS "alias"

		image := fromCmdValue[0]
		alias := fromCmdValue[2]

		if realImage, imageIsAnAlias := aliasImageMap[image]; imageIsAnAlias {
			// "image" is an alias.
			// Bind "alias" to "image"'s real image
			aliasImageMap[alias] = realImage
			continue
		}

		// "image" is a real image
		aliasImageMap[alias] = image
		continue
	}

	if len(lastFromCmd.cmd) == 0 { // No "FROM" command in the Dockerfile
		return "", errors.New("dockerfile must start with a FROM command")
	}

	lastFromValue := lastFromCmd.value[0] // May be a real image or an alias

	if realImage, lastFromIsAnAlias := aliasImageMap[lastFromValue]; lastFromIsAnAlias {
		return realImage, nil
	}

	return lastFromValue, nil
}

func LookupDockerfileLabelValue(
	dockerfilePath string,
	searchedLabelKey string,
) (string, error) {

	dockerfileCmds, err := parseDockerfile(dockerfilePath)

	if err != nil {
		return "", err
	}

	lastLabelValueThatMatch := ""

	for _, dockerfileCmd := range dockerfileCmds {
		if dockerfileCmd.cmd != "LABEL" {
			continue
		}

		labelKeysAndValues := dockerfileCmd.value

		for index := range labelKeysAndValues {

			// The dockerfile parser ensure that
			// all labels have a key and a value so we are
			// guaranteed that all label keys sit at an even index.
			// eg: [labelKey, labelValue, labelKey2, labelValue2, ...]
			isLabelKey := index%2 == 0

			if !isLabelKey {
				continue
			}

			// Label keys and values could be quoted.
			// See: https://docs.docker.com/engine/reference/builder/#label
			labelKey := strings.Trim(labelKeysAndValues[index], "\"'")
			labelValue := strings.Trim(labelKeysAndValues[index+1], "\"'")

			if labelKey == searchedLabelKey {
				lastLabelValueThatMatch = labelValue
			}
		}
	}

	return lastLabelValueThatMatch, nil
}

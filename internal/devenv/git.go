package devenv

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/recode-sh/recode/github"
)

func cloneGitHubRepo(
	repoOwner string,
	repoName string,
	cloneDir string,
) error {

	maxRetries := 3
	retriesInterval := 4 * time.Second
	var lastErrorReturned error

	// Avoid a GitHub race condition that
	// prevent an SSH key pair to be used
	// just after creation
	for retry := 0; retry < maxRetries; retry++ {
		githubGitURL := github.BuildGitURL(repoOwner, repoName)

		cmd := exec.Command(
			"git",
			"clone",
			"--quiet",
			string(githubGitURL),
			cloneDir,
		)

		var stdout bytes.Buffer
		var stderr bytes.Buffer

		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()

		if err != nil {
			newLineRegExp := regexp.MustCompile(`\n+`)

			lastErrorReturned = fmt.Errorf(
				"error while cloning the repository \"%s/%s\".\n\n%s\n\n%s",
				repoOwner,
				repoName,
				strings.TrimSpace(
					newLineRegExp.ReplaceAllLiteralString(stderr.String(), " "),
				),
				err.Error(),
			)

			time.Sleep(retriesInterval)

			continue
		}

		lastErrorReturned = nil
		break
	}

	return lastErrorReturned
}

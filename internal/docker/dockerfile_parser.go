package docker

import (
	"fmt"
	"io"
	"os"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
)

type dockerfileCmd struct {
	cmd       string   // command name (ex: "FROM")
	subCmd    string   // for ONBUILD only this holds the sub-command
	json      bool     // whether the value is written in json form
	original  string   // The original source line
	startLine int      // The original source line number which starts this command
	endLine   int      // The original source line number which ends this command
	flags     []string // Any flags such as "--from=..." for "COPY".
	value     []string // The contents of the command (ex: "ubuntu:xenial")
}

func parseDockerfile(filename string) ([]dockerfileCmd, error) {
	file, err := os.Open(filename)

	if err != nil {
		return nil, fmt.Errorf(
			"error opening Dockerfile \"%s\": %s",
			filename,
			err.Error(),
		)
	}

	defer file.Close()

	dockerfileCMDs, err := parseDockerfileReader(file)

	if err != nil {
		return nil, fmt.Errorf(
			"error parsing Dockerfile \"%s\": %s",
			filename,
			err.Error(),
		)
	}

	return dockerfileCMDs, nil
}

func parseDockerfileReader(dockerfileReader io.Reader) ([]dockerfileCmd, error) {
	dockerfileParsed, err := parser.Parse(dockerfileReader)

	if err != nil {
		return nil, err
	}

	var dockerfileCmds []dockerfileCmd

	for _, child := range dockerfileParsed.AST.Children {
		cmd := dockerfileCmd{
			cmd:       child.Value,
			original:  child.Original,
			startLine: child.StartLine,
			endLine:   child.EndLine,
			flags:     child.Flags,
		}

		// Only happens for "ONBUILD" commands
		if child.Next != nil && len(child.Next.Children) > 0 {
			cmd.subCmd = child.Next.Children[0].Value
			child = child.Next.Children[0]
		}

		cmd.json = child.Attributes["json"]

		for n := child.Next; n != nil; n = n.Next {
			cmd.value = append(cmd.value, n.Value)
		}

		dockerfileCmds = append(dockerfileCmds, cmd)
	}

	return dockerfileCmds, nil
}

package docker

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
)

type BuildOutput struct {
	Stream      string           `json:"stream"`
	Error       string           `json:"error"`
	ErrorDetail BuildErrorDetail `json:"errorDetail"`
}

type BuildErrorDetail struct {
	Message string `json:"message"`
}

func HandleBuildOutput(
	buildOutputReader io.Reader,
	streamHandler func(logLine string) error,
) error {

	scanner := bufio.NewScanner(buildOutputReader)

	for scanner.Scan() {
		buildOutputJSON := scanner.Text()
		buildOutput := &BuildOutput{}

		err := json.Unmarshal([]byte(buildOutputJSON), buildOutput)

		if err != nil {
			return err
		}

		if buildOutput.Error != "" {
			return errors.New(buildOutput.ErrorDetail.Message)
		}

		if buildOutput.Stream == "" {
			continue
		}

		err = streamHandler(
			buildOutput.Stream,
		)

		if err != nil {
			return err
		}
	}

	return scanner.Err()
}

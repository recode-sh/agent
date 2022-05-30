package docker

import (
	"path/filepath"
	"testing"
)

func TestLookupDockerfileBaseImage(t *testing.T) {
	testCases := []struct {
		test              string
		dockerfilePath    string
		expectedBaseImage string
	}{
		{
			test:              "base",
			dockerfilePath:    "base.Dockerfile",
			expectedBaseImage: "recodesh/base-dev-env:latest",
		},

		{
			test:              "multi_from",
			dockerfilePath:    "multi_from.Dockerfile",
			expectedBaseImage: "base",
		},

		{
			test:              "multistages",
			dockerfilePath:    "multi_stages.Dockerfile",
			expectedBaseImage: "alpine:latest",
		},

		{
			test:              "multi_stages_walk",
			dockerfilePath:    "multi_stages_walk.Dockerfile",
			expectedBaseImage: "ubuntu",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.test, func(t *testing.T) {
			baseImage, err := LookupDockerfileBaseImage(
				filepath.Join("./testdata", tc.dockerfilePath),
			)

			if err != nil {
				t.Fatalf("expected no error, got '%+v'", err)
			}

			if baseImage != tc.expectedBaseImage {
				t.Fatalf(
					"expected base image to equal '%s', got '%s'",
					tc.expectedBaseImage,
					baseImage,
				)
			}
		})
	}
}

func TestLookupDockerfileLabelValue(t *testing.T) {
	testCases := []struct {
		test               string
		dockerfilePath     string
		expectedLabelValue string
	}{
		{
			test:               "base",
			dockerfilePath:     "base.Dockerfile",
			expectedLabelValue: "golang.go,dbaeumer.vscode-eslint",
		},

		{
			test:               "multi_from",
			dockerfilePath:     "multi_from.Dockerfile",
			expectedLabelValue: "dbaeumer.vscode-eslint",
		},

		{
			test:               "multistages",
			dockerfilePath:     "multi_stages.Dockerfile",
			expectedLabelValue: "",
		},

		{
			test:               "multi_stages_walk",
			dockerfilePath:     "multi_stages_walk.Dockerfile",
			expectedLabelValue: "dbaeumer.vscode-eslint",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.test, func(t *testing.T) {
			labelValue, err := LookupDockerfileLabelValue(
				filepath.Join("./testdata", tc.dockerfilePath),
				"sh.recode.vscode.extensions",
			)

			if err != nil {
				t.Fatalf("expected no error, got '%+v'", err)
			}

			if labelValue != tc.expectedLabelValue {
				t.Fatalf(
					"expected label value to equal '%s', got '%s'",
					tc.expectedLabelValue,
					labelValue,
				)
			}
		})
	}
}

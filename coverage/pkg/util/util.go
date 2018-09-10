package util

import (
	"fmt"
	"golang.org/x/tools/cover"
	"io"
	"k8s.io/test-infra/coverage/pkg/cov"
	"os"
)

func DumpProfile(destination string, profile []*cover.Profile) error {
	var output io.Writer
	if destination == "-" {
		output = os.Stdout
	} else {
		f, err := os.Create(destination)
		if err != nil {
			return fmt.Errorf("failed to open %s: %v", destination, err)
		}
		defer f.Close()
		output = f
	}
	err := cov.DumpProfile(profile, output)
	if err != nil {
		return fmt.Errorf("failed to dump profile: %v", err)
	}
	return nil
}

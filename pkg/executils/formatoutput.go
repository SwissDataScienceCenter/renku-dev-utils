package executils

import (
	"fmt"
	"os/exec"
)

func FormatOutput(output []byte, err error) ([]byte, error) {
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return output, fmt.Errorf("%s", string(ee.Stderr))
		}
	}
	return output, err
}

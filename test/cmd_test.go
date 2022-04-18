package test

import (
	"os/exec"
	"testing"
)

func Test_cmd_run(t *testing.T) {
	cmd := exec.Command("git", "version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}
	println(string(output))
}

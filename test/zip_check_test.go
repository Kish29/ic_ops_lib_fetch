package test

import (
	"os/exec"
	"testing"
)

func CheckZipIntegrity(filename string) bool {
	err := exec.Command("zip", `-T`, filename).Run()
	return err == nil
}

func Test_zip1(t *testing.T) {
	println(CheckZipIntegrity("../source_code/detools/0.51.0.zip"))
}

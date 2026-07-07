//go:build ignore

package main

import (
	"fmt"
	"os"
	"os/exec"
)

func run(cmd string, args ...string) int {
	fmt.Printf("===== COMMAND: %s =====\n", append([]string{cmd}, args...))
	c := exec.Command(cmd, args...)
	c.Dir = "/Volumes/Satechi/Developer/mockport"
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	err := c.Run()
	code := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			code = exitErr.ExitCode()
		} else {
			code = 1
		}
	}
	fmt.Printf("EXIT_CODE=%d\n\n", code)
	return code
}

func main() {
	run("bash", "scripts/check-doc-links.sh")
	run("bash", "scripts/check-public-trust.sh")
}

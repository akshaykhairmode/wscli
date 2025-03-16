package main

import (
	"fmt"
	"os/exec"
	"testing"
	"time"
)

func getExecuteParams() []string {
	return []string{"run", "main.go", "-c", "ws://localhost:8080/ws", "-w", "5s", "-x", "hello world"}
}

func TestExecute(t *testing.T) {

	now := time.Now()

	cmd := exec.Command("go", getExecuteParams()...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Error(err)
		return
	}

	if time.Since(now) < 5*time.Second {
		t.Error("program did not wait for 5seconds")
	}

	fmt.Println(string(output))

	if string(output) != "« hello world\n" {
		t.Error("output does not match")
	}

}

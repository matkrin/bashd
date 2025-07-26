package server

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"slices"

	"github.com/matkrin/bashd/logger"
)

func getDocumentation(command string) string {
	var documentation string
	if slices.Contains(append(BASH_KEYWORDS[:], BASH_BUILTINS[:]...), command) {
		documentation = runHelp(command)
	} else {
		documentation = runMan(command)
	}

	return fmt.Sprintf("```man\n%s\n```", documentation)
}

func runMan(command string) string {
	manCmd := exec.Command("man", "-p", "cat", command)
	colCmd := exec.Command("col", "-bx")

	man, err := runPipe(manCmd, colCmd)
	if err != nil {
		logger.Errorf("Error running pipe: %s", err)
	}
	return man
}

func runHelp(command string) string {
	helpCmd := exec.Command("bash", "-c", fmt.Sprintf("help %s", command))
	colCmd := exec.Command("col", "-bx")

	help, err := runPipe(helpCmd, colCmd)
	if err != nil {
		logger.Errorf("Error running pipe: %s", err)
	}
	return help
}

func runPipe(cmd1, cmd2 *exec.Cmd) (string, error) {
	pipeReader, pipeWriter := io.Pipe()
	cmd1.Stdout = pipeWriter
	cmd2.Stdin = pipeReader

	var out bytes.Buffer
	cmd2.Stdout = &out

	if err := cmd1.Start(); err != nil {
		return "", fmt.Errorf("Error running command %v", cmd1)
	}
	if err := cmd2.Start(); err != nil {
		return "", fmt.Errorf("Error running command %v", cmd2)
	}

	go func() {
		defer pipeWriter.Close()
		cmd1.Wait()
	}()

	if err := cmd2.Wait(); err != nil {
		return "", fmt.Errorf("Error waiting for command %v", cmd2)
	}

	return out.String(), nil
}

package util

import (
	"bufio"
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"os"
	"path"
	"path/filepath"
	"strings"
	"syscall"
)

var (
	configDir = ""
)

func AddNamespaceToArgs(args []string, namespace string) []string {
	if namespace == "" {
		return args
	}

	return append(args, "--namespace", namespace)
}

func GetRunaiConfigDir() (string, error) {
	if configDir != "" {
		return configDir, nil
	}

	dir, err := os.Executable()
	if err != nil {
		return "", err
	}

	realPath, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return "", err
	}

	return path.Dir(realPath), nil
}

// ReadString reads a string from the stdin.
func ReadString(prompt string) (string, error) {
	if _, err := fmt.Fprint(os.Stderr, prompt); err != nil {
		return "", fmt.Errorf("write error: %v", err)
	}
	r := bufio.NewReader(os.Stdin)
	s, err := r.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("read error: %v", err)
	}
	s = strings.TrimRight(s, "\r\n")
	return s, nil
}

// ReadPassword reads a password from the stdin without echo back.
func ReadPassword(prompt string) (string, error) {
	if _, err := fmt.Fprint(os.Stderr, prompt); err != nil {
		return "", fmt.Errorf("write error: %v", err)
	}
	b, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", fmt.Errorf("read error: %v", err)
	}
	if _, err := fmt.Fprintln(os.Stderr); err != nil {
		return "", fmt.Errorf("write error: %v", err)
	}
	return string(b), nil
}
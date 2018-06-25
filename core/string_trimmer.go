package core

import (
	"os"

	"golang.org/x/crypto/ssh/terminal"
)

func TrimString(input string, prefixLength int) string {
	trimAt, _, _ := terminal.GetSize(int(os.Stdout.Fd()))
	trimAt = trimAt - prefixLength
	if len(input) < trimAt {
		trimAt = len(input)
	}
	return input[0:trimAt]
}

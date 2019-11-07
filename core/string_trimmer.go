package core

import (
	"os"

	"golang.org/x/crypto/ssh/terminal"
)

const DefaultTrimAtLength = 120

func TrimString(input string, extraCharacterLength int) string {
	trimAt, _, err := terminal.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		trimAt = DefaultTrimAtLength
	}
	trimAt = trimAt - extraCharacterLength
	if len(input) < trimAt {
		return input
	}
	return input[0:trimAt] + "\u2026"
}

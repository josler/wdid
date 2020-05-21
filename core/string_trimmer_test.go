package core

import (
	"testing"

	"gotest.tools/assert"
)

func TestStringTrimmerTrimsExtraCharacters(t *testing.T) {
	trimmed := TrimString("my input", DefaultTrimAtLength-len("my"))
	assert.Equal(t, trimmed, "my…")
}

func TestStringTrimmerPreservesNoEllipsisIfInputSmaller(t *testing.T) {
	trimmed := TrimString("small", DefaultTrimAtLength-20)
	assert.Equal(t, trimmed, "small")
}

func TestStringTrimmerNegative(t *testing.T) {
	trimmed := TrimString("", DefaultTrimAtLength+5)
	assert.Equal(t, trimmed, "…")
}

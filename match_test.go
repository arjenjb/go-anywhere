package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_MatchEmptySearch(t *testing.T) {
	assert.True(t, Matches("", "Hallo"))
}

func Test_MatchDifferentPositions(t *testing.T) {
	assert.True(t, Matches("wi", strings.ToLower("whimsical")))
	assert.False(t, Matches("wi", strings.ToLower("incognito")))
}

func Test_MatchLongerSuffix(t *testing.T) {
	assert.True(t, Matches("ial", "mercurial"))
	assert.False(t, Matches("ials", "mercurial"))
}

func Test_MatchShortCircuit(t *testing.T) {
	assert.False(t, Matches("aap", "a"))
	assert.False(t, Matches("abcdefg", "_______abc"))
}

func Test_MatchExact(t *testing.T) {
	assert.True(t, Matches("aap", "aap"))
}

func Test_MatchBoundaries(t *testing.T) {
	assert.True(t, Matches("ae", "abcde"))
}

func Test_MatchScore_Exact(t *testing.T) {
	assert.Equal(t, MatchScore("wi", "wi"), 1.0)
	assert.Equal(t, MatchScore("bc", "abcd"), 0.5)
	assert.Equal(t, MatchScore("abc", "abcd"), 0.75)

	assert.Equal(t, MatchScore("abc", "abcdefghijklmnopqrst"), 0.15)
	assert.Equal(t, MatchScore("acd", "abcdefghijklmnopqrst"), 0.08625)
	assert.Equal(t, MatchScore("ace", "abcdefghijklmnopqrst"), 0.06499999999999999)

	assert.Equal(t, MatchScore("a", "abcd"), 0.25)

	assert.Equal(t, MatchScore("x", "y"), 0.0)

	print(MatchScore("abcdefghijlmnop", "MaxxAudio Pro by Waves - Audio instelling voor luidsprekers"))
}

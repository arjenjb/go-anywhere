package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseLinkHeader(t *testing.T) {
	links, err := ParseLinkHeader("<https://api.github.com/user/repos?page=2>; rel=\"next\", <https://api.github.com/user/repos?page=6>; rel=\"last\"\n")

	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, "https://api.github.com/user/repos?page=2", links[0].Url)
	assert.Equal(t, "next", links[0].Rel)

	assert.Equal(t, "https://api.github.com/user/repos?page=6", links[1].Url)
	assert.Equal(t, "last", links[1].Rel)
}

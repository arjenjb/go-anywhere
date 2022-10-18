package api

import (
	"fmt"
	"net/http"

	"github.com/davecgh/go-spew/spew"
	"github.com/google/go-querystring/query"
)

type GithubRestClient struct {
	Token string
}

func (c *GithubRestClient) ListRepositoriesForAuthenticatedUser() (it *PagingIterator[Repository]) {
	return &PagingIterator[Repository]{
		i:        -1,
		nextLink: c.url("/user/repos"),
		client:   c,
	}
}

func (c *GithubRestClient) ListMatchingRefs(r ListMatchingRefsRequest) (it *Iterator[Reference]) {
	return &Iterator[Reference]{
		url:    c.url(fmt.Sprintf("/repos/%s/git/matching-refs/%s", r.Repo, r.Ref)),
		client: c,
	}
}

func (c *GithubRestClient) ListPulls(r ListPullsRequest) (it *PagingIterator[Pull]) {
	v, _ := query.Values(r)
	path := fmt.Sprintf("/repos/%s/pulls?%s", r.Repo, v.Encode())

	return &PagingIterator[Pull]{
		i:        -1,
		nextLink: c.url(path),
		client:   c,
	}
}

func (c *GithubRestClient) url(s string) string {
	return fmt.Sprintf(fmt.Sprintf("https://api.github.com%s", s))
}

func (c *GithubRestClient) executeUrl(u string) (*http.Response, error) {
	req, _ := http.NewRequest("GET", u, nil)
	req.Header.Set("Authorization", spew.Sprintf("token %s", c.Token))
	req.Header.Set("Accept", "application/vnd.github+json")

	return http.DefaultClient.Do(req)
}

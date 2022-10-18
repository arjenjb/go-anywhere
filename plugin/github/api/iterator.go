package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
)

type Iterator[T any] struct {
	url     string
	fetched bool
	i       int
	items   []T
	client  *GithubRestClient
}

func (r *Iterator[T]) Next() (ok bool, item T, err error) {
	if !r.fetched {
		err = r.fetch()
		if err != nil {
			return
		}
	}

	r.i++
	ok = r.i < len(r.items)

	if ok {
		item = r.items[r.i]
	}

	return
}

func (r *Iterator[T]) fetch() error {
	resp, err := r.client.executeUrl(r.url)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("Failed to make request: %s", resp.Status))
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	items := make([]T, 0)

	err = json.Unmarshal(data, &items)
	if err != nil {
		return err
	}

	r.fetched = true
	r.items = items
	r.i = -1

	return nil
}

type PagingIterator[T any] struct {
	items    []T
	i        int
	nextLink string

	client *GithubRestClient
}

func (p *PagingIterator[T]) atEndOfPage() bool {
	return p.i == len(p.items)-1
}

func (p *PagingIterator[T]) hasNextPage() bool {
	return len(p.nextLink) > 0
}

func (p *PagingIterator[T]) Next() (found bool, item T, err error) {
	if p.atEndOfPage() {
		if !p.hasNextPage() {
			found = false
			return
		}

		log.Println("[INFO] Fetching next page")
		err = p.fetchNext()
		if err != nil {
			found = false
			return
		}
	}

	p.i++

	found = len(p.items) > p.i

	if found {
		item = p.items[p.i]
	}

	return
}

func (p *PagingIterator[T]) fetchNext() error {
	resp, err := p.client.executeUrl(p.nextLink)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("Failed to make request: %s", resp.Status))
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	items := make([]T, 0)

	err = json.Unmarshal(data, &items)
	if err != nil {
		return err
	}

	links, err := ParseLinkHeader(resp.Header.Get("Link"))
	if err != nil {
		return err
	}

	if link, found := links.FindByRel("next"); found {
		p.nextLink = link.Url
	} else {
		// No more pages, clear it
		p.nextLink = ""
	}

	p.items = items
	p.i = -1

	return nil
}

package api

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

var token string

func init() {
	godotenv.Load("../../../.env")
	token = os.Getenv("GITHUB_TOKEN")
}

func requireToken(t *testing.T) {
	if token == "" {
		t.Skip("GITHUB_TOKEN not set in .env")
	}
}

func TestGithubRestClient_ListRepositories(t *testing.T) {
	requireToken(t)

	c := GithubRestClient{Token: token}
	it := c.ListRepositoriesForAuthenticatedUser()

	for {
		found, next, err := it.Next()

		if err != nil {
			log.Printf("[ERROR] %s", err)
		}

		if !found {
			break
		}

		fmt.Println(next.FullName)

	}
}

func TestGithubRestClient_ListPulls(t *testing.T) {
	requireToken(t)

	c := GithubRestClient{Token: token}
	it := c.ListPulls(ListPullsRequest{
		Repo:    "arjenjb/go-anywhere",
		State:   Open,
		PerPage: 2,
		Page:    1,
	})

	for {
		found, pr, err := it.Next()

		if err != nil {
			log.Printf("[ERROR] %s", err)
		}

		if !found {
			break
		}

		fmt.Println("--------")
		fmt.Println(pr.Title)
		fmt.Println(pr.User.Login)
	}
}

func TestGithubRestClient_ListMatchingRefs(t *testing.T) {
	requireToken(t)

	c := GithubRestClient{Token: token}
	it := c.ListMatchingRefs(ListMatchingRefsRequest{
		Repo: "arjenjb/go-anywhere",
		Ref:  "tags/",
	})

	for {
		found, pr, err := it.Next()

		if err != nil {
			log.Printf("[ERROR] %s", err)
		}

		if !found {
			break
		}

		fmt.Println("--------")
		fmt.Println(pr.Ref)

	}
}

package github

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"image"
	"log"
	"time"

	"go-keyboard-launcher/api"
	github "go-keyboard-launcher/plugin/github/api"
)

//go:embed logo.png
var iconData []byte

const (
	BranchesCategory    = api.User + 1
	PullRequestCategory = api.User + 2
	TagsCategory        = api.User + 3
)

const (
	KeywordConfigure uint8 = iota
)

type Config struct {
	Token string
}

type Plugin struct {
	icon *image.Image

	config       Config
	state        []string
	repositories []api.Item
	client       *github.GithubRestClient
}

func (p *Plugin) Name() string {
	return "github"
}

func (p *Plugin) LoadConfig(load func(interface{}) error) {
	if err := load(&p.config); err != nil {
		fmt.Printf("Failed to load github config")
	} else {
		fmt.Printf("Config loaded")

		p.client = &github.GithubRestClient{Token: p.config.Token}
	}
}

func (p *Plugin) Initialize() {
	decoded, _, err := image.Decode(bytes.NewReader(iconData))
	if err == nil {
		p.icon = &decoded
	}
}

func (p *Plugin) Catalog() error {
	result := make([]api.Item, 0)

	it := p.client.ListRepositoriesForAuthenticatedUser()

	for {
		found, repo, err := it.Next()

		if err != nil {
			return err
		} else if !found {
			break
		}

		result = append(result, api.Item{
			Label:       repo.FullName,
			Description: repo.Description,
			Category:    api.Url,
			Target:      repo.HtmlUrl,
			ArgsHint:    api.Accepted,
			Data:        repo.FullName,
		})
	}

	p.repositories = result
	return nil
}

func (p *Plugin) Icon() *image.Image {
	return p.icon
}

func (p *Plugin) GetItems() ([]api.Item, error) {
	return p.repositories, nil
}

func (p *Plugin) Execute(item api.Item) {
	log.Printf("I don't know how to execute item %s", item.String())
}

func (p *Plugin) Suggest(ctx context.Context, input string, chain []api.Item, setSuggestions api.SuggestionCallback) {
	if len(chain) == 0 {
		return
	} else if len(chain) == 1 {

		setSuggestions([]api.Item{
			{
				Label:    fmt.Sprintf("Pull requests"),
				Category: PullRequestCategory,
				Target:   fmt.Sprintf("%s/pulls", chain[0].Target),
				ArgsHint: api.Accepted,
			},
			{
				Label:    fmt.Sprintf("Tags"),
				Category: TagsCategory,
				ArgsHint: api.Required,
			},
			{
				Label:    fmt.Sprintf("Branches"),
				Category: BranchesCategory,
				ArgsHint: api.Required,
			},
		}, api.MatchFuzzy)
	} else if len(chain) == 2 {
		repoName := chain[0].Data.(string)

		switch chain[1].Category {
		case TagsCategory:
			it := p.client.ListMatchingRefs(github.ListMatchingRefsRequest{
				Repo: repoName,
				Ref:  "tags/",
			})

			var suggestions []api.Item

			for {
				ok, item, err := it.Next()
				if err != nil {
					log.Printf("[Error] %s", err)
					return
				} else if !ok {
					break
				}

				suggestions = append(suggestions, api.Item{
					Label:    item.Ref,
					Category: api.Url,
					Target:   item.Url,
					ArgsHint: api.Forbidden,
				})
			}

			setSuggestions(suggestions, api.MatchFuzzy)

		case PullRequestCategory:
			it := p.client.ListPulls(github.ListPullsRequest{
				Repo:    repoName,
				State:   github.Open,
				PerPage: 30,
				Page:    1,
			})

			var suggestions []api.Item
			now := time.Now()

			for {
				ok, item, err := it.Next()
				if err != nil {
					log.Printf("[Error] %s", err)
					return
				} else if !ok {
					break
				}

				delta := humanReadableTimeDelta(now.Sub(item.CreatedAt))

				suggestions = append(suggestions, api.Item{
					Label:       item.Title,
					Description: fmt.Sprintf("#%d  opened %s  by %s", item.Number, delta, item.User.Login),
					Category:    api.Url,
					Target:      item.HtmlUrl,
					ArgsHint:    api.Forbidden,
				})
			}

			setSuggestions(suggestions, api.MatchFuzzy)
		}
	}
}

func humanReadableTimeDelta(sub time.Duration) string {

	printUnit := func(n int, unit string) string {
		if n == 1 {
			return fmt.Sprintf("%d %s ago", n, unit)
		} else {
			return fmt.Sprintf("%d %ss ago", n, unit)
		}
	}

	hours := sub.Hours()
	if hours < 1.0 {
		return printUnit(int(hours*60.0), "minute")
	} else if hours < 24 {
		return printUnit(int(hours), "hour")
	} else if hours < (30 * 24) {
		return printUnit(int(hours/30.0), "day")
	} else {
		n := time.Now()
		t := n.Add(-sub)

		if n.Year() == t.Year() {
			return fmt.Sprintf("on %s", t.Format("Jan _2"))
		} else {
			return fmt.Sprintf("on %s", t.Format("Jan _2 2006"))
		}
	}
}

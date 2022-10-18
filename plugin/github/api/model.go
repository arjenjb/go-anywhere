package api

import "time"

type Owner struct {
	Login string `json:"login"`
}

type Pull struct {
	Url               string `json:"url"`
	HtmlUrl           string `json:"html_url"`
	DiffUrl           string `json:"diff_url"`
	PatchUrl          string `json:"patch_url"`
	IssueUrl          string `json:"issue_url"`
	CommitsUrl        string `json:"commits_url"`
	CommentsUrl       string `json:"comments_url"`
	ReviewCommentsUrl string `json:"review_comments_url"`
	StatusesUrl       string `json:"statuses_url"`

	Id     int       `json:"id"`
	NodeId string    `json:"node_id"`
	Number int       `json:"number"`
	State  PullState `json:"state"`
	Locked bool      `json:"locked"`
	Title  string    `json:"title"`
	User   Owner     `json:"user"`
	Body   string    `json:"body"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	ClosedAt  time.Time `json:"closed_at"`
	MergedAt  time.Time `json:"merged_at"`
}

type Repository struct {
	Id            int    `json:"id"`
	Name          string `json:"name"`
	FullName      string `json:"full_name"`
	Owner         Owner  `json:"owner"`
	Private       bool   `json:"private"`
	HtmlUrl       string `json:"html_url"`
	Description   string `json:"description"`
	Url           string `json:"url"`
	DefaultBranch string `json:"default_branch"`
	Archived      bool   `json:"archived"`
	Disabled      bool   `json:"disabled"`
	Visibility    string `json:"visibility"`
	PushedAt      string `json:"pushed_at"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

type Object struct {
	Type string `json:"type"`
	Sha  string `json:"sha"`
	Url  string `json:"url"`
}

type Reference struct {
	Ref    string `json:"ref"`
	NodeId string `json:"node_id"`
	Url    string `json:"url"`
	Object Object `json:"object"`
}

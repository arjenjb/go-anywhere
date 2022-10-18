package api

type SortDirection string

const (
	Asc  SortDirection = "asc"
	Desc               = "desc"
)

type PullState string

const (
	Open   PullState = "open"
	Closed           = "closed"
	All              = "all"
)

type ListMatchingRefsRequest struct {
	Repo string `url:"-"`
	Ref  string `url:"-"`
}

type ListPullsRequest struct {
	Repo      string         `url:"-"`
	State     PullState      `url:"state"`
	Head      *string        `url:"head,omitempty"`
	Base      *string        `url:"base,omitempty"`
	Sort      *string        `url:"sort,omitempty"`
	Direction *SortDirection `url:"direction,omitempty"`
	PerPage   int            `url:"per_page"`
	Page      int            `url:"page"`
}

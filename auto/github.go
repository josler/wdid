package auto

import (
	"context"
	"fmt"
	"time"

	_github "github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type GithubClient struct {
	ctx      context.Context
	username string
	client   *_github.Client
}

func NewGithubClient(ctx context.Context, username, accessToken string) *GithubClient {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)
	tc := oauth2.NewClient(context.TODO(), ts)
	return &GithubClient{
		ctx:      ctx,
		username: username,
		client:   _github.NewClient(tc),
	}
}

func (ghc *GithubClient) isVerbose() bool {
	v := ghc.ctx.Value("verbose")
	return v != nil && v.(bool)
}

func (ghc *GithubClient) Load(startTime time.Time, endTime time.Time) []*Option {
	options := []*Option{}

	// https://developer.github.com/v3/search/#search-issues
	queryString := fmt.Sprintf("involves:%s updated:>%s", ghc.username, startTime.Format("2006-01-02T15:04:05"))

	if ghc.isVerbose() {
		fmt.Println(queryString)
	}
	issueSearch, _, err := ghc.client.Search.Issues(context.TODO(), queryString, &_github.SearchOptions{Sort: "created", ListOptions: _github.ListOptions{PerPage: 100}})

	if err != nil {
		fmt.Println(err)
		return options
	}
	for _, issue := range issueSearch.Issues {
		issueType := "issue"
		if issue.IsPullRequest() {
			issueType = "pr"
		}
		status := "waiting"
		if !issue.GetClosedAt().IsZero() {
			status = "done"
		}
		opt := Option{
			data:     fmt.Sprintf("[%s] @%s %s %s", issueType, *issue.User.Login, *issue.Title, issue.GetHTMLURL()),
			dateTime: *issue.UpdatedAt,
			status:   status,
		}
		options = append(options, &opt)
	}
	return options
}

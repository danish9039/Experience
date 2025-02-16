package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/danish9039/Experience/Jaeger"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

const (
	owner          = "jaegertracing"
	repo           = "jaeger"
	repoIDL        = "jaeger-idl"
	username       = "danish9039"
	readmeDataPath = "./Jaeger/readmeData.go"
	mdFileName     = "./Jaeger/README.md"
)

func fetchPRs(client *github.Client, ctx context.Context, owner, repo string) ([]*github.PullRequest, error) {
	prs := make([]*github.PullRequest, 0)
	for i := 1; i <= 30; i++ {
		opts := &github.PullRequestListOptions{
			State:       "closed",
			Head:        username,
			ListOptions: github.ListOptions{PerPage: 100, Page: i},
		}
		pr, _, err := client.PullRequests.List(ctx, owner, repo, opts)
		if err != nil {
			return nil, fmt.Errorf("error fetching pull requests: %v", err)
		}
		prs = append(prs, pr...)
		fmt.Println("Fetching PRs from", repo, "Page:", i)
	}
	return prs, nil
}

func main() {
	ctx := context.Background()
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		fmt.Println("Set the GITHUB_TOKEN environment variable.")
		return
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	// Fetch PRs from both repositories
	prsJaeger, err := fetchPRs(client, ctx, owner, repo)
	if err != nil {
		fmt.Println(err)
		return
	}

	prsIDL, err := fetchPRs(client, ctx, owner, repoIDL)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Combine PRs from both repositories
	allPRs := append(prsJaeger, prsIDL...)

	filteredPRs := make([]*github.PullRequest, 0)
	for _, pull := range allPRs {
		if *pull.User.Login == username {
			filteredPRs = append(filteredPRs, pull)
		}
	}

	var sb strings.Builder
	sb.WriteString(Jaeger.Data)
	sb.WriteString("\n\n")

	sb.WriteString("| Date Created | Title | Pull Request Link |\n")
	sb.WriteString("| ------------ | ----- | ----------------- |\n")

	for _, pr := range filteredPRs {
		date := pr.CreatedAt.Format(time.DateOnly)
		title := strings.ReplaceAll(*pr.Title, "|", "\\|")
		url := *pr.HTMLURL
		sb.WriteString(fmt.Sprintf("| %s | %s | [PR link](%s) |\n", date, title, url))
	}

	err = os.WriteFile(mdFileName, []byte(sb.String()), 0644)
	if err != nil {
		fmt.Printf("Error writing markdown file: %v\n", err)
		return
	}

	fmt.Printf("Markdown file '%s' created.\n", mdFileName)
}

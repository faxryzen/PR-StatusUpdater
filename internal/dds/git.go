package dds

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type PullRequest struct {
	Number  uint                 `json:"number"`
	Author  string               `json:"author"`
	LabID   string               `json:"labID"`
	Times   map[string]time.Time `json:"times"` //created fined merged
	Marks   []string             `json:"marks"`
	Score   int                  `json:"score"`
	Debug   string               `json:"debug"`
}

func toMoscow(t time.Time) time.Time {
	loc, _ := time.LoadLocation("Europe/Moscow")
	return t.In(loc)
}

type Queries struct {
	MergedPRsQuery string
	OpenPRsQuery   string
}

func GetGraphQLForGit(repo Repository) (Queries, error) {
	merged, err := os.ReadFile(queryFilepath + "pr_merged.graphql")
	if err != nil {
		return Queries{}, fmt.Errorf("failed to read merged_prs.graphql: %w", err)
	}

	open, err := os.ReadFile(queryFilepath + "pr_open.graphql")
	if err != nil {
		return Queries{}, fmt.Errorf("failed to read open_prs.graphql: %w", err)
	}

	mergedQuery := strings.ReplaceAll(string(merged), "$owner", repo.Auth)
	mergedQuery = strings.ReplaceAll(mergedQuery, "$name", repo.Name)
	
	openQuery := strings.ReplaceAll(string(open), "$owner", repo.Auth)
	openQuery = strings.ReplaceAll(openQuery, "$name", repo.Name)

	return Queries{string(mergedQuery), string(openQuery)}, nil
}

type Repository struct {
	Name string
	Auth string
}

func GetRepositories() ([]Repository, error) {
	data, err := os.ReadFile(configFilepath + "repos.csv")
	if err != nil {
		return nil, err
	}

	repos := []Repository{}
	records := strings.Split(string(data), "\n")
	for _, repo := range records {
		repoInfo := strings.Split(repo, ",")
		if len(repoInfo) == 2 {
			repos = append(repos, Repository{
				Name: repoInfo[0],
				Auth: repoInfo[1],
			})
		}
	}

	return repos, nil
}

func GetPullRequests(query string) ([]PullRequest, error) {
	jqFilter, err := os.ReadFile(filterFilepath + "pr_filter.jq")
	if err != nil {
		return nil, fmt.Errorf("failed to read jq filter: %w", err)
	}

	cmd := exec.Command(
		"gh", "api", "graphql",
		"-f", "query="+query,
		"--jq", string(jqFilter))

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	if len(out) == 0 {
		return []PullRequest{}, nil
	}

	var pullRequests []PullRequest
	lines := strings.Split(string(out), "\n")

	for _, line := range lines {

		if line == "" {
			continue
		}

		var pr PullRequest

		err = json.Unmarshal([]byte(line), &pr)
		if err != nil {
			panic("what da hell")
		}

		if pr.LabID == "" {
			continue
		}

		pr.Times["created"] = toMoscow(pr.Times["created"])

		if pr.Times["fined"].IsZero() {
			delete(pr.Times, "fined")
		} else {
			pr.Times["fined"] = toMoscow(pr.Times["fined"])
		}
		if pr.Times["merged"].IsZero() {
			delete(pr.Times, "merged")
		} else {
			pr.Times["merged"] = toMoscow(pr.Times["merged"])
		}

		pr.Score = -1

		pullRequests = append(pullRequests, pr)
	}

	return pullRequests, nil
}